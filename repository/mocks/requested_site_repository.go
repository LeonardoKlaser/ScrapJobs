package mocks

import "github.com/stretchr/testify/mock"

type MockRequestedSiteRepository struct {
	mock.Mock
}

func (m *MockRequestedSiteRepository) Create(userID int, url string) error {
	args := m.Called(userID, url)
	return args.Error(0)
}
