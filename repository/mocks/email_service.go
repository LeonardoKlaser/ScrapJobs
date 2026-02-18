package mocks

import (
	"context"
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error {
	args := m.Called(ctx, userEmail, job, analysis)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, userEmail, userName, dashboardLink string) error {
	args := m.Called(ctx, userEmail, userName, dashboardLink)
	return args.Error(0)
}
