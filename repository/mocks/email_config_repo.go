package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockEmailConfigRepo struct {
	mock.Mock
}

func (m *MockEmailConfigRepo) GetAll() ([]model.EmailProviderConfig, error) {
	args := m.Called()
	return args.Get(0).([]model.EmailProviderConfig), args.Error(1)
}

func (m *MockEmailConfigRepo) Update(configs []model.EmailProviderConfig, updatedBy int) error {
	args := m.Called(configs, updatedBy)
	return args.Error(0)
}
