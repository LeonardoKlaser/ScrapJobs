package interfaces

import "web-scrapper/model"

type NotificationRepositoryInterface interface {
	InsertNewNotification(jobId int, userId int) error
	GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error)
	GetNotificationsByUser(userId int, limit int) ([]model.NotificationWithJob, error)
	GetMonthlyAnalysisCount(userID int) (int, error)
}
