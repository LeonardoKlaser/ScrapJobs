package interfaces

type NotificationRepositoryInterface interface {
	InsertNewNotification(jobId int, userId int) error
	GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error)
}
