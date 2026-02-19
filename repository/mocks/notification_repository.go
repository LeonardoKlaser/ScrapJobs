package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) InsertNewNotification(jobId int, userId int) error {
	args := m.Called(jobId, userId)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error) {
	args := m.Called(userId, jobs)
	return args.Get(0).(map[int]bool), args.Error(1)
}

func (m *MockNotificationRepository) GetNotificationsByUser(userId int, limit int) ([]model.NotificationWithJob, error) {
	args := m.Called(userId, limit)
	return args.Get(0).([]model.NotificationWithJob), args.Error(1)
}

func (m *MockNotificationRepository) GetMonthlyAnalysisCount(userID int) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}
