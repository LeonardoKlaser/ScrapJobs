package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"web-scrapper/model"
	"web-scrapper/repository"
)

type AnalysisService interface {
    Analyze(ctx context.Context, curriculum model.Curriculum, job model.Job) (model.ResumeAnalysis, error)
}

type EmailService interface {
    SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error
}

type NotificationsUsecase struct{
	userSiteRepo repository.UserSiteRepository
	curriculumRepo repository.CurriculumRepository
	analysisService AnalysisService
	emailService EmailService
	notificationRepository repository.NotificationRepository
}

func NewNotificationUsecase(
	userSiteRepo repository.UserSiteRepository,
    curriculumRepo repository.CurriculumRepository,
    analysisService AnalysisService,
    emailService EmailService,
	notificationRepository repository.NotificationRepository,
) *NotificationsUsecase{
	return &NotificationsUsecase{
		userSiteRepo:    userSiteRepo,
        curriculumRepo:  curriculumRepo,
        analysisService: analysisService,
        emailService:    emailService,
		notificationRepository: notificationRepository,
	}
}

func (s *NotificationsUsecase) FindMatchesAndNotify(siteId int, jobs []*model.Job) error{
	log.Printf("Jobs: %v", jobs)
	userWithCurriculum, err := s.userSiteRepo.GetUsersBySiteId(siteId)
	if err != nil {
		return fmt.Errorf("error to get users by site Id %d: %w", siteId, err)
	}

	jobsById := make(map[int]*model.Job)
	for _, job := range jobs {
		jobsById[job.ID] = job
	}

	for _, user := range userWithCurriculum{

		if user.Curriculum == nil {
			log.Printf("INFO: User %s (ID: %d) has no curriculum, skipping notifications for this user.", user.Name, user.UserId)
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
			return err
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

		//inicializa worker pool para melhorar performance (executar mais de uma notificação/analise ao mesmo tempo)
		const numWorkers = 5
		jobsChan := make(chan *model.Job, len(jobsToSend))
		var wg sync.WaitGroup

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int){
				defer wg.Done()

				for job := range jobsChan{
					log.Printf("actual job: %v", job)
					analysis, err := s.analysisService.Analyze(context.Background(), *user.Curriculum, *job)
					if err != nil {
						log.Printf("[Worker %d] ERROR: AI analysis failed for job %s: %v", workerID, job.Title, err)
						continue 
					}
					log.Printf("analise feita para usuario %s com job %s", user.Name, job.Title)

					err = s.emailService.SendAnalysisEmail(context.Background(), user.Email, *job, analysis)
					if err != nil {
						log.Printf("[Worker %d] ERROR: Email sending failed for job %s: %v", workerID, job.Title, err)
						continue
					}
					log.Printf("email feita para usuario %s com job %s", user.Name, job.Title)
					err = s.notificationRepository.InsertNewNotification(job.ID, user.UserId)
					if err != nil {
						log.Printf("[Worker %d] FATAL: Failed to insert notification record for job %d: %v", workerID, job.ID, err)
					}
				}
			}(i)
		}

		for _, job := range jobsToSend {
			jobsChan <- job
		}
		close(jobsChan) 

		wg.Wait()
		log.Printf("Finished notification pool for user %s.", user.Name)

	}
	return nil
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