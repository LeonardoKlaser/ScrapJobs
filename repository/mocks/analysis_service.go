package mocks

import (
	"context"
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockAnalysisService struct{
	mock.Mock
}

func (m *MockAnalysisService) Analyze(ctx context.Context ,curriculum model.Curriculum, job model.Job) (model.ResumeAnalysis, error) {
	args := m.Called(ctx, curriculum, job)
	return args.Get(0).(model.ResumeAnalysis), args.Error(1)
}