package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockUserSiteRepository struct{
	mock.Mock
}

func (m *MockUserSiteRepository) GetUsersBySiteId(siteId int) ([]model.UserSiteCurriculum, error){
	args := m.Called(siteId)
	return args.Get(0).([]model.UserSiteCurriculum), args.Error(1)
}


func (m *MockUserSiteRepository) InsertNewUserSite(userId int, siteId int, filters []string) error{
	args := m.Called(userId, siteId, filters)
	return args.Error(0)
}