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
}

func NewNotificationUsecase(
	userSiteRepo interfaces.UserSiteRepositoryInterface,
    analysisService interfaces.AnalysisService,
    emailService interfaces.EmailService,
	notificationRepository interfaces.NotificationRepositoryInterface,
	asynqClient *asynq.Client,
	planRepository interfaces.PlanRepositoryInterface,
) *NotificationsUsecase{
	return &NotificationsUsecase{
		userSiteRepo:    userSiteRepo,
        analysisService: analysisService,
        emailService:    emailService,
		notificationRepository: notificationRepository,
		asynqClient: asynqClient,
		planRepository: planRepository,
	}
}

func (s *NotificationsUsecase) FindMatches(siteId int, jobs []*model.Job) ([]tasks.NotifyNewJobsPayload, error) {
	var payloadsToEnqueue []tasks.NotifyNewJobsPayload
	userWithCurriculum, err := s.userSiteRepo.GetUsersBySiteId(siteId)
	if err != nil {
		return payloadsToEnqueue, fmt.Errorf("error to get users by site Id %d: %w", siteId, err)
	}

	jobsById := make(map[int]*model.Job)
	for _, job := range jobs {
		jobsById[job.ID] = job
	}

	for _, user := range userWithCurriculum {
		var jobsToSend []*model.Job
		var matchedJobIDs []int
		for _, job := range jobs {
			if s.matchJobWithFilters(*job, user.TargetWords) {
				matchedJobIDs = append(matchedJobIDs, job.ID)
			}
		}

		if len(matchedJobIDs) == 0 {
			continue
		}
		logging.Logger.Info().Str("user_name", user.Name).Msg("User matched for notification")
		notifiedJobsMap, err := s.notificationRepository.GetNotifiedJobIDsForUser(user.UserId, matchedJobIDs)
		if err != nil {
			return payloadsToEnqueue, err
		}

		for _, jobId := range matchedJobIDs {
			if _, alreadyNotified := notifiedJobsMap[jobId]; !alreadyNotified {
				jobsToSend = append(jobsToSend, jobsById[jobId])
			}
		}

		logging.Logger.Debug().Int("user_id", user.UserId).Int("job_count", len(jobsToSend)).Msg("Jobs list created for notification")
		if len(jobsToSend) == 0 {
			continue
		}

		payloadsToEnqueue = append(payloadsToEnqueue, tasks.NotifyNewJobsPayload{
			User: user,
			Jobs: jobsToSend,
		})
	}
	return payloadsToEnqueue, nil
}

func (s *NotificationsUsecase) matchJobWithFilters(job model.Job, filters []string) bool {
    if len(filters) == 0 {
        return true 
    }

    
    jobTitleLower := strings.ToLower(job.Title)

    for _, filter := range filters {
        if strings.Contains(jobTitleLower, strings.ToLower(filter)) {
            return true 
        }
    }

    return false 
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

func (s *NotificationsUsecase) ProcessNewJobsNotification(ctx context.Context, user model.UserSiteCurriculum, jobs []*model.Job) error {
	// Insert notification records for all jobs first (idempotency)
	for _, job := range jobs {
		err := s.notificationRepository.InsertNewNotification(job.ID, user.UserId)
		if err != nil {
			logging.Logger.Error().Err(err).Int("job_id", job.ID).Int("user_id", user.UserId).Msg("Failed to insert notification record")
		}
	}

	err := s.emailService.SendNewJobsEmail(ctx, user.Email, user.Name, jobs)
	if err != nil {
		return fmt.Errorf("ERROR: Email sending failed for user %s: %w", user.Name, err)
	}
	logging.Logger.Info().Str("user_name", user.Name).Int("job_count", len(jobs)).Msg("New jobs email sent")
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