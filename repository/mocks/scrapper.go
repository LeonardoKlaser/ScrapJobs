package mocks

import (
	"context"
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockScraper struct {
	mock.Mock
}

func (m *MockScraper) Scrape(ctx context.Context, config model.SiteScrapingConfig) ([]*model.Job, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Job), args.Error(1)
}
