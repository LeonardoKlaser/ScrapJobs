package usecase

import (
	"context"
	"web-scrapper/interfaces"
	"web-scrapper/model"

	"golang.org/x/time/rate"
)


type RateLimitedAiAnalyser struct {
	next    interfaces.AnalysisService
	limiter *rate.Limiter
}


func NewRateLimitedAiAnalyser(next interfaces.AnalysisService, limiter *rate.Limiter) interfaces.AnalysisService {
	return &RateLimitedAiAnalyser{
		next:    next,
		limiter: limiter,
	}
}

func (r *RateLimitedAiAnalyser) Analyze(ctx context.Context, curriculum model.Curriculum, job model.Job) (model.ResumeAnalysis, error) {
	
	err := r.limiter.Wait(ctx)
	if err != nil {
		return model.ResumeAnalysis{}, err
	}
	return r.next.Analyze(ctx, curriculum, job)
}