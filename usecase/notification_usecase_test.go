package usecase

import (
	"context"
	"fmt"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationsUsecase_MatchJobsForUser(t *testing.T) {
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockEmailService := new(mocks.MockEmailService)
	mockPlanRepo := new(mocks.MockPlanRepository)

	notificationUsecase := NewNotificationUsecase(
		mockUserSiteRepo,
		nil,
		mockEmailService,
		mockNotificationRepo,
		mockPlanRepo,
		nil,
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

func TestNotificationsUsecase_SendDigestForUser(t *testing.T) {
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockEmailService := new(mocks.MockEmailService)
	mockUserRepo := new(mocks.MockUserRepository)

	notificationUsecase := NewNotificationUsecase(
		nil,
		nil,
		mockEmailService,
		mockNotificationRepo,
		nil,
		mockUserRepo,
	)

	t.Run("should send digest email and mark notifications as SENT", func(t *testing.T) {
		userID := 10
		pendingJobs := []model.NotificationWithJob{
			{ID: 1, JobID: 100, UserID: userID, JobTitle: "Go Dev", JobCompany: "Acme", JobLocation: "Remote", JobLink: "https://acme.com/1"},
			{ID: 2, JobID: 101, UserID: userID, JobTitle: "Go Senior", JobCompany: "Acme", JobLocation: "SP", JobLink: "https://acme.com/2"},
		}

		mockNotificationRepo.On("GetPendingJobsForUser", userID).Return(pendingJobs, nil).Once()
		mockUserRepo.On("GetUserBasicInfo", userID).Return("Test User", "test@example.com", nil).Once()
		mockEmailService.On("SendNewJobsEmail", mock.Anything, "test@example.com", "Test User", mock.AnythingOfType("[]*model.Job")).Return(nil).Once()
		mockNotificationRepo.On("BulkUpdateNotificationStatus", userID, []int{100, 101}, "SENT").Return(nil).Once()

		err := notificationUsecase.SendDigestForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockEmailService.AssertExpectations(t)
	})

	t.Run("should return nil when no pending notifications exist", func(t *testing.T) {
		userID := 20

		mockNotificationRepo.On("GetPendingJobsForUser", userID).Return([]model.NotificationWithJob{}, nil).Once()

		err := notificationUsecase.SendDigestForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockEmailService.AssertNotCalled(t, "SendNewJobsEmail")
	})

	t.Run("should NOT mark SENT if email fails", func(t *testing.T) {
		userID := 30
		pendingJobs := []model.NotificationWithJob{
			{ID: 3, JobID: 200, UserID: userID, JobTitle: "Dev", JobCompany: "Beta", JobLocation: "RJ", JobLink: "https://beta.com/1"},
		}

		mockNotificationRepo.On("GetPendingJobsForUser", userID).Return(pendingJobs, nil).Once()
		mockUserRepo.On("GetUserBasicInfo", userID).Return("Fail User", "fail@example.com", nil).Once()
		mockEmailService.On("SendNewJobsEmail", mock.Anything, "fail@example.com", "Fail User", mock.AnythingOfType("[]*model.Job")).Return(fmt.Errorf("SES error")).Once()

		err := notificationUsecase.SendDigestForUser(context.Background(), userID)

		assert.Error(t, err)
		mockNotificationRepo.AssertNotCalled(t, "BulkUpdateNotificationStatus")
	})
}
