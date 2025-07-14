package usecase

import (
	"context"
	"log"
	"fmt"
	"strings"
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
	userWithCurriculum, err := s.userSiteRepo.GetUsersBySiteId(siteId)
	if err != nil {
		return fmt.Errorf("error to get users by site Id %d: %w", siteId, err)
	}

	jobsById := make(map[int]*model.Job)
	for _, job := range jobs {
		jobsById[job.ID] = job
	}

	for _, user := range userWithCurriculum{
		var matchedJobIDs []int
		for _, job := range jobs{
			if s.matchJobWithFilters(*job, user.TargetWords){
				matchedJobIDs = append(matchedJobIDs, job.ID)
			}
		}
		notifiedJobsMap, err := s.notificationRepository.GetNotifiedJobIDsForUser(user.UserId, matchedJobIDs)
		if err != nil{
			return err
		}

		for _, jobId := range  matchedJobIDs {
			if _, alreadyNotified := notifiedJobsMap[jobId]; !alreadyNotified {
				job := jobsById[jobId]

				analysis, err := s.analysisService.Analyze(context.Background(), *user.Curriculum, *job)
				if err != nil {
					log.Printf("error to get AI analysis for user: %s about vacancy: %s: %v", user.Name, job.Title, err)
					continue
				}

				err = s.emailService.SendAnalysisEmail(context.Background(), user.Email, *job, analysis)

				if err != nil{
					log.Printf("error to send email notification for %s about %s : %v", user.Email, job.Title, err)
				}

				err = s.notificationRepository.InsertNewNotification(job.ID, user.UserId)
                if err != nil {
                     log.Printf("FATAL: could not insert notification record for user %d, job %d: %v", user.UserId, job.ID, err)
                }
			}
		}
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