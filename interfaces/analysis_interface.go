package interfaces

import (
	"context"
	"web-scrapper/model"
)

type AnalysisService interface {
	Analyze(ctx context.Context, curriculum model.Curriculum, job model.Job) (model.ResumeAnalysis, error)
}
