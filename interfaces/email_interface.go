package interfaces

import (
	"context"
	"web-scrapper/model"
)

type EmailService interface {
	SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error
	SendWelcomeEmail(ctx context.Context, userEmail, userName, dashboardLink string) error
	SendNewJobsEmail(ctx context.Context, userEmail string, userName string, jobs []*model.Job) error
	SendPasswordResetEmail(ctx context.Context, email, userName, resetLink string) error
}
