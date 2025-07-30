package usecase

import (
	"errors"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationsUsecase_FindMatchesAndNotify(t *testing.T){
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockAnalysisService := new(mocks.MockAnalysisService)
	mockEmailService := new(mocks.MockEmailService)
    clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: "redis:6379"})
    defer clientAsynq.Close()

	notificationUsecase := NewNotificationUsecase(
		mockUserSiteRepo,
		mockAnalysisService,
		mockEmailService,
		mockNotificationRepo,
        clientAsynq,
	)

	siteId := 1
	testUser := model.UserSiteCurriculum{
		UserId:      10,
        Name:        "Test User",
        Email:       "test@example.com",
        Curriculum:  &model.Curriculum{Skills: "Go"},
        TargetWords: []string{"developer"},
	}

	matchingJob := &model.Job{ID: 101, Title: "Go Developer"}
	nonMatchingJob := &model.Job{ID: 102, Title: "Python Engineer"}

	t.Run("Should send notification for a new matching job", func(t *testing.T) {
		mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
        mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(make(map[int]bool), nil).Once()
        mockAnalysisService.On("Analyze", mock.Anything, *testUser.Curriculum, *matchingJob).Return(model.ResumeAnalysis{}, nil).Once()
        mockEmailService.On("SendAnalysisEmail", mock.Anything, testUser.Email, *matchingJob, mock.Anything).Return(nil).Once()
        mockNotificationRepo.On("InsertNewNotification", matchingJob.ID, testUser.UserId).Return(nil).Once()

		_,err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob, nonMatchingJob})

		assert.NoError(t, err)
		mockUserSiteRepo.AssertExpectations(t)
        mockNotificationRepo.AssertExpectations(t)
        mockAnalysisService.AssertExpectations(t)
        mockEmailService.AssertExpectations(t)
	})

	t.Run("should NOT send notification for an already notified job", func(t *testing.T) {
        alreadyNotifiedMap := map[int]bool{matchingJob.ID: true}
        mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
        mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(alreadyNotifiedMap, nil).Once()

        
        _,err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob})

        // Asserções
        assert.NoError(t, err)
        mockUserSiteRepo.AssertExpectations(t)
        mockNotificationRepo.AssertExpectations(t)
    })


	t.Run("should continue processing even if one email fails", func(t *testing.T) {
        mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
        mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(make(map[int]bool), nil).Once()
        mockAnalysisService.On("Analyze", mock.Anything, *testUser.Curriculum, *matchingJob).Return(model.ResumeAnalysis{}, nil).Once()
        
        mockEmailService.On("SendAnalysisEmail", mock.Anything, testUser.Email, *matchingJob, mock.Anything).Return(errors.New("aws ses failed")).Once()
        

       
        _, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob})

        
        assert.NoError(t, err) 
        mockUserSiteRepo.AssertExpectations(t)
        mockNotificationRepo.AssertExpectations(t)
        mockAnalysisService.AssertExpectations(t)
        mockEmailService.AssertExpectations(t)
    })


}