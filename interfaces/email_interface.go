package interfaces

import(
	"web-scrapper/model"
	"context"
)

type EmailService interface {
    SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error
}