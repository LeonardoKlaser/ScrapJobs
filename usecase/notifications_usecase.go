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

func (s *NotificationsUsecase) FindMatches(siteId int, jobs []*model.Job) ([]tasks.AnalyzeUserJobPayload, error){
	var payloadsToEnqueue []tasks.AnalyzeUserJobPayload
	userWithCurriculum, err := s.userSiteRepo.GetUsersBySiteId(siteId)
	if err != nil {
		return payloadsToEnqueue, fmt.Errorf("error to get users by site Id %d: %w", siteId, err)
	}

	jobsById := make(map[int]*model.Job)
	for _, job := range jobs {
		jobsById[job.ID] = job
	}

	for _, user := range userWithCurriculum{

		if user.Curriculum == nil {
			logging.Logger.Info().Str("user_name", user.Name).Int("user_id", user.UserId).Msg("User has no curriculum, skipping")
			continue
		}

		// Verificar limite de análises do plano
		plan, planErr := s.planRepository.GetPlanByUserID(user.UserId)
		if planErr != nil {
			logging.Logger.Error().Err(planErr).Int("user_id", user.UserId).Msg("Error fetching user plan, skipping user")
			continue
		}
		monthlyCount, countErr := s.notificationRepository.GetMonthlyAnalysisCount(user.UserId)
		if countErr != nil {
			logging.Logger.Error().Err(countErr).Int("user_id", user.UserId).Msg("Error fetching monthly analysis count, skipping user")
			continue
		}
		if plan == nil {
			logging.Logger.Warn().Int("user_id", user.UserId).Msg("User has no plan assigned, skipping limit check")
		}
		// MaxAIAnalyses <= 0 = unlimited
		if plan != nil && plan.MaxAIAnalyses > 0 && monthlyCount >= plan.MaxAIAnalyses {
			logging.Logger.Info().Int("user_id", user.UserId).Int("count", monthlyCount).Int("limit", plan.MaxAIAnalyses).Msg("User reached AI analysis limit, skipping")
			continue
		}
		remainingQuota := -1 // unlimited
		if plan != nil && plan.MaxAIAnalyses > 0 {
			remainingQuota = plan.MaxAIAnalyses - monthlyCount
		}

		var jobsToSend []*model.Job
		var matchedJobIDs []int
		for _, job := range jobs{
			if s.matchJobWithFilters(*job, user.TargetWords){
				matchedJobIDs = append(matchedJobIDs, job.ID)
			}
		}

		if len(matchedJobIDs) == 0 {
			continue 
		}
		logging.Logger.Info().Str("user_name", user.Name).Msg("User matched for notification")
		notifiedJobsMap, err := s.notificationRepository.GetNotifiedJobIDsForUser(user.UserId, matchedJobIDs)
		if err != nil{
			return payloadsToEnqueue, err
		}

		for _, jobId := range  matchedJobIDs {
			if _, alreadyNotified := notifiedJobsMap[jobId]; !alreadyNotified {
				jobsToSend = append(jobsToSend, jobsById[jobId])
			}
		}

		logging.Logger.Debug().Int("user_id", user.UserId).Int("job_count", len(jobsToSend)).Msg("Jobs list created for notification")
		if len(jobsToSend) == 0 {
			continue
		}

		// Truncar para o limite restante do plano
		if remainingQuota >= 0 && len(jobsToSend) > remainingQuota {
			jobsToSend = jobsToSend[:remainingQuota]
		}
		if len(jobsToSend) == 0 {
			continue
		}

		for _, job := range jobsToSend{
			payloadsToEnqueue = append(payloadsToEnqueue, tasks.AnalyzeUserJobPayload{
				User: user,
				Job: job,
			})
		}

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

func (s *NotificationsUsecase) ProcessingJobAnalyze(ctx context.Context, job model.Job, user model.UserSiteCurriculum) (tasks.NotifyUserPayload, error){
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