package interfaces

type RequestedSiteRepositoryInterface interface {
	Create(userID int, url string) error
}
