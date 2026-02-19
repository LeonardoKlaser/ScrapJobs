package usecase

import (
	"context"
	"fmt"
	"log"
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
}

func NewNotificationUsecase(
	userSiteRepo interfaces.UserSiteRepositoryInterface,
    analysisService interfaces.AnalysisService,
    emailService interfaces.EmailService,
	notificationRepository interfaces.NotificationRepositoryInterface,
	asynqClient *asynq.Client,
) *NotificationsUsecase{
	return &NotificationsUsecase{
		userSiteRepo:    userSiteRepo,
        analysisService: analysisService,
        emailService:    emailService,
		notificationRepository: notificationRepository,
		asynqClient: asynqClient,
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
		log.Printf("usuario encontrando para enviar notificacao: %s", user.Name)
		notifiedJobsMap, err := s.notificationRepository.GetNotifiedJobIDsForUser(user.UserId, matchedJobIDs)
		if err != nil{
			return payloadsToEnqueue, err
		}

		for _, jobId := range  matchedJobIDs {
			if _, alreadyNotified := notifiedJobsMap[jobId]; !alreadyNotified {
				jobsToSend = append(jobsToSend, jobsById[jobId])
			}
		}

		log.Printf("Criado lista de jobs para notificar")
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
	log.Printf("analise feita para usuario %s com job %s", user.Name, job.Title)
		
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

	err := s.emailService.SendAnalysisEmail(ctx, user.Email, job, analysis)
	if err != nil {
		return fmt.Errorf("ERROR: Email sending failed for job %s: %v", job.Title, err)			
	}
	log.Printf("email feita para usuario %s com job %s", user.Name, job.Title)
	err = s.notificationRepository.InsertNewNotification(job.ID, user.UserId)
	if err != nil {
		return fmt.Errorf("FATAL: Failed to insert notification record for job %d: %v", job.ID, err)
	}

	return nil
}