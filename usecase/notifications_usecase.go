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
}

func NewNotificationUsecase(
	userSiteRepo repository.UserSiteRepository,
    curriculumRepo repository.CurriculumRepository,
    analysisService AnalysisService,
    emailService EmailService,
) *NotificationsUsecase{
	return &NotificationsUsecase{
		userSiteRepo:    userSiteRepo,
        curriculumRepo:  curriculumRepo,
        analysisService: analysisService,
        emailService:    emailService,
	}
}

func (s *NotificationsUsecase) FindMatchesAndNotify(siteId int, jobs []*model.Job) error{
	users, err := s.userSiteRepo.GetUsersBySiteId(siteId)
	if err != nil {
		return fmt.Errorf("error to get users by site Id %d: %w", siteId, err)
	}

	for _, user := range users{
		for _, job := range jobs{
			if s.matchJobWithFilters(*job, user.TargetWords){
				curriculum, err := s.curriculumRepo.FindCurriculumByUserID(user.UserId)
				if err != nil {
					log.Printf("error to get user %s curriculum: %v", user.Name, err)
					continue	
				}

				analysis, err := s.analysisService.Analyze(context.Background(), curriculum, *job)
				if err != nil {
					log.Printf("error to get AI analysis for user: %s about vacancy: %s: %v", user.Name, job.Title, err)
					continue
				}

				err = s.emailService.SendAnalysisEmail(context.Background(), user.Email, *job, analysis)

				if err != nil{
					log.Printf("error to send email notification for %s about %s : %v", user.Email, job.Title, err)
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