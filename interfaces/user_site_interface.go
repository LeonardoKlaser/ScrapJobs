package interfaces

import "web-scrapper/model"

type UserSiteRepositoryInterface interface {
	GetUsersBySiteId(siteId int) ([]model.UserSiteCurriculum, error)
	InsertNewUserSite(userId int, siteId int, filters []string) error
}