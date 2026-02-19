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

func (m *MockUserSiteRepository) GetSubscribedSiteIDs(userId int) (map[int]bool, error) {
	args := m.Called(userId)
	return args.Get(0).(map[int]bool), args.Error(1)
}

func (m *MockUserSiteRepository) DeleteUserSite(userId int, siteId string) error {
	args := m.Called(userId, siteId)
	return args.Error(0)
}

func (m *MockUserSiteRepository) UpdateUserSiteFilters(userId int, siteId int, filters []string) error {
	args := m.Called(userId, siteId, filters)
	return args.Error(0)
}

func (m *MockUserSiteRepository) GetUserSiteCount(userID int) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}