package usecase

import (
	"context"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestNotificationsUsecase_FindMatchesAndNotify(t *testing.T) {
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

	matchingJob := &model.Job{ID: 101, Title: "Go Developer"}
	nonMatchingJob := &model.Job{ID: 102, Title: "Python Engineer"}

	t.Run("Should return payload for a new matching job", func(t *testing.T) {
		mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
		mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(make(map[int]bool), nil).Once()

		payloads, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob, nonMatchingJob})

		assert.NoError(t, err)
		assert.Len(t, payloads, 1)
		assert.Len(t, payloads[0].Jobs, 1)
		assert.Equal(t, matchingJob.ID, payloads[0].Jobs[0].ID)
		assert.Equal(t, testUser.UserId, payloads[0].User.UserId)
		mockUserSiteRepo.AssertExpectations(t)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should NOT return payload for an already notified job", func(t *testing.T) {
		alreadyNotifiedMap := map[int]bool{matchingJob.ID: true}
		mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{testUser}, nil).Once()
		mockNotificationRepo.On("GetNotifiedJobIDsForUser", testUser.UserId, []int{matchingJob.ID}).Return(alreadyNotifiedMap, nil).Once()

		payloads, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob})

		assert.NoError(t, err)
		assert.Len(t, payloads, 0)
		mockUserSiteRepo.AssertExpectations(t)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should return payload even without curriculum (no AI needed)", func(t *testing.T) {
		userNoCurriculum := model.UserSiteCurriculum{
			UserId:      20,
			Name:        "No Curriculum User",
			Email:       "nocv@example.com",
			Curriculum:  nil,
			TargetWords: []string{"developer"},
		}
		mockUserSiteRepo.On("GetUsersBySiteId", siteId).Return([]model.UserSiteCurriculum{userNoCurriculum}, nil).Once()
		mockNotificationRepo.On("GetNotifiedJobIDsForUser", userNoCurriculum.UserId, []int{matchingJob.ID}).Return(make(map[int]bool), nil).Once()

		payloads, err := notificationUsecase.FindMatches(siteId, []*model.Job{matchingJob})

		assert.NoError(t, err)
		assert.Len(t, payloads, 1)
		assert.Len(t, payloads[0].Jobs, 1)
		mockUserSiteRepo.AssertExpectations(t)
		mockNotificationRepo.AssertExpectations(t)
	})
}

func TestNotificationsUsecase_MatchJobsForUser(t *testing.T) {
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockEmailService := new(mocks.MockEmailService)
	mockPlanRepo := new(mocks.MockPlanRepository)
	clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: "redis:6379"})
	defer clientAsynq.Close()

	notificationUsecase := NewNotificationUsecase(
		mockUserSiteRepo,
		nil,
		mockEmailService,
		mockNotificationRepo,
		clientAsynq,
		mockPlanRepo,
	)

	t.Run("should bulk insert PENDING for matching jobs", func(t *testing.T) {
		userID := 10
		jobsWithFilters := []model.JobWithFilters{
			{JobID: 1, Title: "Go Developer", Company: "Acme", Filters: []string{"developer"}},
			{JobID: 2, Title: "Python Engineer", Company: "Acme", Filters: []string{"developer"}},
			{JobID: 3, Title: "Go Developer Senior", Company: "Beta", Filters: []string{"developer"}},
		}

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return(jobsWithFilters, nil).Once()
		mockNotificationRepo.On("BulkInsertPendingNotifications", userID, []int{1, 3}).Return(nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should skip bulk insert when no jobs match filters", func(t *testing.T) {
		userID := 20
		jobsWithFilters := []model.JobWithFilters{
			{JobID: 5, Title: "Python Engineer", Company: "Acme", Filters: []string{"java"}},
		}

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return(jobsWithFilters, nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertNotCalled(t, "BulkInsertPendingNotifications")
	})

	t.Run("should match all jobs when user has no filters", func(t *testing.T) {
		userID := 30
		jobsWithFilters := []model.JobWithFilters{
			{JobID: 10, Title: "Any Job", Company: "Acme", Filters: []string{}},
			{JobID: 11, Title: "Another Job", Company: "Beta", Filters: []string{}},
		}

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return(jobsWithFilters, nil).Once()
		mockNotificationRepo.On("BulkInsertPendingNotifications", userID, []int{10, 11}).Return(nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should return nil when no unnotified jobs exist", func(t *testing.T) {
		userID := 40

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return([]model.JobWithFilters{}, nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertNotCalled(t, "BulkInsertPendingNotifications")
	})
}
