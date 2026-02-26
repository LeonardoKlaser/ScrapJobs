package mocks

import (
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) CreateJob(job model.Job) (int, error) {
	args := m.Called(job)
	return args.Int(0), args.Error(1)
}

func (m *MockJobRepository) FindJobByRequisitionID(requisition_ID int) (bool, error) {
	args := m.Called(requisition_ID)
	return args.Bool(0), args.Error(1)
}

func (m *MockJobRepository) FindJobsByRequisitionIDs(requisition_IDs []int64) (map[int64]bool, error) {
	args := m.Called(requisition_IDs)
	return args.Get(0).(map[int64]bool), args.Error(1)
}

func (m *MockJobRepository) UpdateLastSeen(requisition_ID int64) (int, error) {
	args := m.Called(requisition_ID)
	return args.Int(0), args.Error(1)
}

func (m *MockJobRepository) DeleteOldJobs() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockJobRepository) GetJobByID(jobID int) (*model.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}
