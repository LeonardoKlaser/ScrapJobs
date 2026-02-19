package usecase

import (
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestNotificationsUsecase_FindMatchesAndNotify(t *testing.T){
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockAnalysisService := new(mocks.MockAnalysisService)
	mockEmailService := new(mocks.MockEmailService)
	mockPlanRepo := new(mocks.MockPlanRepository)
    clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: "redis:6379"})
    defer clientAsynq.Close()

	notificationUsecase := NewNotificationUsecase(
		mockUserSiteRepo,
		mockAnalysisService,
		mockEmailService,
		mockNotificationRepo,
        clientAsynq,
		mockPlanRepo,
	)

	siteId := 1
	testUser := model.UserSiteCurriculum{
		UserId:      10,
        Name:        "Test User",
        Email:       "test@example.com",
        Curriculum:  &model.Curriculum{Skills: "Go"},
        TargetWords: []string{"developer"},
	}

	// Plan com limite generoso (100 análises) para os testes
	testPlan := &model.Plan{ID: 1, Name: "Profissional", MaxAIAnalyses: 100, MaxSites: 15}

	matchingJob := &model.Job{ID: 101, Title: "Go Developer"}
	nonMatchingJob := &model.Job{ID: 102, Title: "Python Engineer"}

	t.Run("Should return payload for a new matching job", func(t *testing.T) {
		mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
		mockPlanRepo.On("GetPlanByUserID", testUser.UserId).Return(testPlan, nil).Once()
		mockNotificationRepo.On("GetMonthlyAnalysisCount", testUser.UserId).Return(0, nil).Once()
        mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(make(map[int]bool), nil).Once()

		payloads, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob, nonMatchingJob})

		assert.NoError(t, err)
		assert.Len(t, payloads, 1)
		assert.Equal(t, matchingJob.ID, payloads[0].Job.ID)
		assert.Equal(t, testUser.UserId, payloads[0].User.UserId)
		mockUserSiteRepo.AssertExpectations(t)
        mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should NOT return payload for an already notified job", func(t *testing.T) {
        alreadyNotifiedMap := map[int]bool{matchingJob.ID: true}
        mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
		mockPlanRepo.On("GetPlanByUserID", testUser.UserId).Return(testPlan, nil).Once()
		mockNotificationRepo.On("GetMonthlyAnalysisCount", testUser.UserId).Return(0, nil).Once()
        mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(alreadyNotifiedMap, nil).Once()

        payloads, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob})

        assert.NoError(t, err)
		assert.Len(t, payloads, 0)
        mockUserSiteRepo.AssertExpectations(t)
        mockNotificationRepo.AssertExpectations(t)
    })

	t.Run("should skip user when AI analysis limit is reached", func(t *testing.T) {
		limitedPlan := &model.Plan{ID: 1, Name: "Beta Tester", MaxAIAnalyses: 10, MaxSites: 3}
		mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
		mockPlanRepo.On("GetPlanByUserID", testUser.UserId).Return(limitedPlan, nil).Once()
		mockNotificationRepo.On("GetMonthlyAnalysisCount", testUser.UserId).Return(10, nil).Once()

		payloads, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob})

		assert.NoError(t, err)
		assert.Len(t, payloads, 0)
		mockUserSiteRepo.AssertExpectations(t)
		mockPlanRepo.AssertExpectations(t)
		mockNotificationRepo.AssertExpectations(t)
	})
}
