package mocks

import(
	"github.com/stretchr/testify/mock"
)

type MockNotificationRepository struct{
	mock.Mock
}

func (m *MockNotificationRepository) GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error){
	args := m.Called(userId, jobs)
	return args.Get(0).(map[int]bool), args.Error(1)
}

func (m *MockNotificationRepository) InsertNewNotification(jobId int, userId int) error{
	args := m.Called(jobId, userId)
	return args.Error(0)
}