package usecase

import (
	"errors"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestPlanUsecase_GetAllPlans(t *testing.T) {
	mockRepo := new(mocks.MockPlanRepository)
	uc := NewPlanUsecase(mockRepo)

	t.Run("should return all plans", func(t *testing.T) {
		expected := []model.Plan{
			{ID: 1, Name: "Iniciante", Price: 0, MaxSites: 3, MaxAIAnalyses: 10, Features: pq.StringArray{"feature1"}},
			{ID: 2, Name: "Profissional", Price: 29.90, MaxSites: 15, MaxAIAnalyses: 100, Features: pq.StringArray{"feature1", "feature2"}},
		}

		mockRepo.On("GetAllPlans").Return(expected, nil).Once()

		result, err := uc.GetAllPlans()

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Iniciante", result[0].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repo fails", func(t *testing.T) {
		mockRepo.On("GetAllPlans").Return([]model.Plan{}, errors.New("db error")).Once()

		_, err := uc.GetAllPlans()

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestPlanUsecase_GetPlanByUserID(t *testing.T) {
	mockRepo := new(mocks.MockPlanRepository)
	uc := NewPlanUsecase(mockRepo)

	t.Run("should return plan for user", func(t *testing.T) {
		expected := &model.Plan{ID: 1, Name: "Iniciante", MaxSites: 3}

		mockRepo.On("GetPlanByUserID", 1).Return(expected, nil).Once()

		result, err := uc.GetPlanByUserID(1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Iniciante", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return nil when user has no plan", func(t *testing.T) {
		mockRepo.On("GetPlanByUserID", 99).Return(nil, nil).Once()

		result, err := uc.GetPlanByUserID(99)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestPlanUsecase_GetPlanByID(t *testing.T) {
	mockRepo := new(mocks.MockPlanRepository)
	uc := NewPlanUsecase(mockRepo)

	t.Run("should return plan by ID", func(t *testing.T) {
		expected := &model.Plan{ID: 2, Name: "Profissional", MaxSites: 15}

		mockRepo.On("GetPlanByID", 2).Return(expected, nil).Once()

		result, err := uc.GetPlanByID(2)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Profissional", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return nil when plan not found", func(t *testing.T) {
		mockRepo.On("GetPlanByID", 999).Return(nil, nil).Once()

		result, err := uc.GetPlanByID(999)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
