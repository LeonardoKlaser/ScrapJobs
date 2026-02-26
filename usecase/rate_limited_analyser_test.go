package usecase

import (
	"context"
	"errors"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/time/rate"
)

func TestRateLimitedAiAnalyser_Analyze(t *testing.T) {
	t.Run("should delegate to next analyser successfully", func(t *testing.T) {
		mockNext := new(mocks.MockAnalysisService)
		limiter := rate.NewLimiter(rate.Limit(10), 10)
		uc := NewRateLimitedAiAnalyser(mockNext, limiter)

		curriculum := model.Curriculum{Id: 1, Title: "My CV", Skills: "Go"}
		job := model.Job{ID: 1, Title: "Go Dev"}
		expected := model.ResumeAnalysis{
			MatchAnalysis: model.MatchAnalysis{
				OverallScoreNumeric:     85,
				OverallScoreQualitative: "Bom",
				Summary:                 "Good match",
			},
		}

		mockNext.On("Analyze", mock.Anything, curriculum, job).Return(expected, nil).Once()

		result, err := uc.Analyze(context.Background(), curriculum, job)

		assert.NoError(t, err)
		assert.Equal(t, 85, result.MatchAnalysis.OverallScoreNumeric)
		mockNext.AssertExpectations(t)
	})

	t.Run("should return error when context is cancelled", func(t *testing.T) {
		mockNext := new(mocks.MockAnalysisService)
		limiter := rate.NewLimiter(rate.Limit(0.001), 0) // very low rate to force wait
		uc := NewRateLimitedAiAnalyser(mockNext, limiter)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		curriculum := model.Curriculum{Id: 1}
		job := model.Job{ID: 1}

		_, err := uc.Analyze(ctx, curriculum, job)

		assert.Error(t, err)
		mockNext.AssertNotCalled(t, "Analyze")
	})

	t.Run("should propagate error from next analyser", func(t *testing.T) {
		mockNext := new(mocks.MockAnalysisService)
		limiter := rate.NewLimiter(rate.Limit(10), 10)
		uc := NewRateLimitedAiAnalyser(mockNext, limiter)

		curriculum := model.Curriculum{Id: 1}
		job := model.Job{ID: 1}

		mockNext.On("Analyze", mock.Anything, curriculum, job).Return(model.ResumeAnalysis{}, errors.New("AI error")).Once()

		_, err := uc.Analyze(context.Background(), curriculum, job)

		assert.Error(t, err)
		assert.Equal(t, "AI error", err.Error())
		mockNext.AssertExpectations(t)
	})
}
