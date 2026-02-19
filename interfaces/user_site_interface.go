package interfaces

import "web-scrapper/model"

type UserSiteRepositoryInterface interface {
	GetUsersBySiteId(siteId int) ([]model.UserSiteCurriculum, error)
	InsertNewUserSite(userId int, siteId int, filters []string) error
	GetSubscribedSiteIDs(userId int) (map[int]bool, error)
	DeleteUserSite(userId int, siteId string) error
	UpdateUserSiteFilters(userId int, siteId int, filters []string) error
}
