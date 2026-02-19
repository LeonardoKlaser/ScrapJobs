package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockPlanRepository struct {
	mock.Mock
}

func (m *MockPlanRepository) GetAllPlans() ([]model.Plan, error) {
	args := m.Called()
	return args.Get(0).([]model.Plan), args.Error(1)
}

func (m *MockPlanRepository) GetPlanByUserID(userID int) (*model.Plan, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Plan), args.Error(1)
}

func (m *MockPlanRepository) GetPlanByID(planId int) (*model.Plan, error) {
	args := m.Called(planId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Plan), args.Error(1)
}
