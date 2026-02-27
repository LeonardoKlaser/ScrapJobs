package interfaces

import "web-scrapper/model"

type NotificationRepositoryInterface interface {
	InsertNewNotification(jobId int, userId int) error
	GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error)
	GetNotificationsByUser(userId int, limit int) ([]model.NotificationWithJob, error)
	GetMonthlyAnalysisCount(userID int) (int, error)
	BulkInsertPendingNotifications(userID int, jobIDs []int) error
	GetUserIDsWithPendingNotifications() ([]int, error)
	GetPendingJobsForUser(userID int) ([]model.NotificationWithJob, error)
	BulkUpdateNotificationStatus(userID int, jobIDs []int, status string) error
	GetUnnotifiedJobsForUser(userID int) ([]model.JobWithFilters, error)
	InsertNotificationWithAnalysis(jobId int, userId int, curriculumId int, analysisResult []byte) error
	GetAnalysisHistory(userId int, jobId int) ([]byte, *int, error)
}
