package usecase

import (
	"context"
	"fmt"
	"strings"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/tasks"

	"github.com/hibiken/asynq"
)


type NotificationsUsecase struct{
	userSiteRepo interfaces.UserSiteRepositoryInterface
	analysisService interfaces.AnalysisService
	emailService interfaces.EmailService
	notificationRepository interfaces.NotificationRepositoryInterface
	asynqClient *asynq.Client
	planRepository interfaces.PlanRepositoryInterface
	userRepository interfaces.UserRepositoryInterface
}

func NewNotificationUsecase(
	userSiteRepo interfaces.UserSiteRepositoryInterface,
    analysisService interfaces.AnalysisService,
    emailService interfaces.EmailService,
	notificationRepository interfaces.NotificationRepositoryInterface,
	asynqClient *asynq.Client,
	planRepository interfaces.PlanRepositoryInterface,
	userRepository interfaces.UserRepositoryInterface,
) *NotificationsUsecase{
	return &NotificationsUsecase{
		userSiteRepo:    userSiteRepo,
        analysisService: analysisService,
        emailService:    emailService,
		notificationRepository: notificationRepository,
		asynqClient: asynqClient,
		planRepository: planRepository,
		userRepository: userRepository,
	}
}

func matchJobWithFiltersFromList(jobTitle string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	jobTitleLower := strings.ToLower(jobTitle)
	for _, filter := range filters {
		if strings.Contains(jobTitleLower, strings.ToLower(filter)) {
			return true
		}
	}
	return false
}

func (s *NotificationsUsecase) MatchJobsForUser(ctx context.Context, userID int) error {
	jobs, err := s.notificationRepository.GetUnnotifiedJobsForUser(userID)
	if err != nil {
		return fmt.Errorf("error fetching unnotified jobs for user %d: %w", userID, err)
	}

	if len(jobs) == 0 {
		logging.Logger.Debug().Int("user_id", userID).Msg("No unnotified jobs found for user")
		return nil
	}

	var matchedJobIDs []int
	for _, job := range jobs {
		if matchJobWithFiltersFromList(job.Title, job.Filters) {
			matchedJobIDs = append(matchedJobIDs, job.JobID)
		}
	}

	if len(matchedJobIDs) == 0 {
		logging.Logger.Debug().Int("user_id", userID).Msg("No jobs matched user filters")
		return nil
	}

	if err := s.notificationRepository.BulkInsertPendingNotifications(userID, matchedJobIDs); err != nil {
		return fmt.Errorf("error inserting pending notifications for user %d: %w", userID, err)
	}

	logging.Logger.Info().Int("user_id", userID).Int("matched_count", len(matchedJobIDs)).Msg("Pending notifications created for user")
	return nil
}

func (s *NotificationsUsecase) ProcessingJobAnalyze(ctx context.Context, job model.Job, user model.UserSiteCurriculum) (tasks.NotifyUserPayload, error) {
	analysis, err := s.analysisService.Analyze(ctx, *user.Curriculum, job)
	if err != nil {
		return tasks.NotifyUserPayload{}, fmt.Errorf("ERROR: AI analysis failed for job %s: %v", job.Title, err)				
	}
	logging.Logger.Info().Str("user_name", user.Name).Str("job_title", job.Title).Msg("AI analysis completed")
		
	payload:= tasks.NotifyUserPayload{
		User: user,
		Job: &job,
		Analysis: analysis,
	}
	return payload, nil
}


// GetNotificationsByUser retorna o histórico de notificações de um usuário com dados da vaga
func (s *NotificationsUsecase) GetNotificationsByUser(userId int, limit int) ([]model.NotificationWithJob, error) {
	notifications, err := s.notificationRepository.GetNotificationsByUser(userId, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching notifications for user %d: %w", userId, err)
	}
	return notifications, nil
}

func (s *NotificationsUsecase) ProcessingSingleNotification(ctx context.Context, job model.Job, user model.UserSiteCurriculum, analysis model.ResumeAnalysis) error {
	// Insert notification first to prevent duplicate emails on retry
	err := s.notificationRepository.InsertNewNotification(job.ID, user.UserId)
	if err != nil {
		return fmt.Errorf("FATAL: Failed to insert notification record for job %d: %w", job.ID, err)
	}

	err = s.emailService.SendAnalysisEmail(ctx, user.Email, job, analysis)
	if err != nil {
		return fmt.Errorf("ERROR: Email sending failed for job %s: %w", job.Title, err)
	}
	logging.Logger.Info().Str("user_name", user.Name).Str("job_title", job.Title).Msg("Analysis email sent")

	return nil
}

func (s *NotificationsUsecase) SendDigestForUser(ctx context.Context, userID int) error {
	pendingNotifications, err := s.notificationRepository.GetPendingJobsForUser(userID)
	if err != nil {
		return fmt.Errorf("error fetching pending notifications for user %d: %w", userID, err)
	}

	if len(pendingNotifications) == 0 {
		logging.Logger.Debug().Int("user_id", userID).Msg("No pending notifications for user")
		return nil
	}

	userName, userEmail, err := s.userRepository.GetUserBasicInfo(userID)
	if err != nil {
		return fmt.Errorf("error fetching user info for user %d: %w", userID, err)
	}

	jobs := make([]*model.Job, len(pendingNotifications))
	jobIDs := make([]int, len(pendingNotifications))
	for i, n := range pendingNotifications {
		jobs[i] = &model.Job{
			ID:       n.JobID,
			Title:    n.JobTitle,
			Company:  n.JobCompany,
			Location: n.JobLocation,
			JobLink:  n.JobLink,
		}
		jobIDs[i] = n.JobID
	}

	if err := s.emailService.SendNewJobsEmail(ctx, userEmail, userName, jobs); err != nil {
		return fmt.Errorf("error sending digest email for user %d: %w", userID, err)
	}

	if err := s.notificationRepository.BulkUpdateNotificationStatus(userID, jobIDs, "SENT"); err != nil {
		return fmt.Errorf("error marking notifications as SENT for user %d: %w", userID, err)
	}

	logging.Logger.Info().Int("user_id", userID).Int("job_count", len(jobs)).Msg("Digest email sent and notifications marked as SENT")
	return nil
}