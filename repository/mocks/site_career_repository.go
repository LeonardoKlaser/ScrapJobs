package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockSiteCareerRepository struct {
	mock.Mock
}

func (m *MockSiteCareerRepository) InsertNewSiteCareer(site model.SiteScrapingConfig) (model.SiteScrapingConfig, error) {
	args := m.Called(site)
	return args.Get(0).(model.SiteScrapingConfig), args.Error(1)
}

func (m *MockSiteCareerRepository) GetAllSites() ([]model.SiteScrapingConfig, error) {
	args := m.Called()
	return args.Get(0).([]model.SiteScrapingConfig), args.Error(1)
}
