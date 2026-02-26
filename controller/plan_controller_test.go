package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupPlanController() (*PlanController, *mocks.MockPlanRepository) {
	mockRepo := new(mocks.MockPlanRepository)
	uc := usecase.NewPlanUsecase(mockRepo)
	ctrl := NewPlanController(uc)
	return ctrl, mockRepo
}

func TestPlanController_GetAllPlans(t *testing.T) {
	t.Run("should return all plans successfully", func(t *testing.T) {
		ctrl, mockRepo := setupPlanController()

		expected := []model.Plan{
			{ID: 1, Name: "Iniciante", Price: 0, MaxSites: 3, MaxAIAnalyses: 10, Features: pq.StringArray{"feature1"}},
			{ID: 2, Name: "Profissional", Price: 29.90, MaxSites: 15, MaxAIAnalyses: 100, Features: pq.StringArray{"feature1", "feature2"}},
		}

		mockRepo.On("GetAllPlans").Return(expected, nil).Once()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.GET("/plans", ctrl.GetAllPlans)

		req := httptest.NewRequest("GET", "/plans", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []model.Plan
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result, 2)
		assert.Equal(t, "Iniciante", result[0].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 500 when repo fails", func(t *testing.T) {
		ctrl, mockRepo := setupPlanController()

		mockRepo.On("GetAllPlans").Return([]model.Plan{}, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.GET("/plans", ctrl.GetAllPlans)

		req := httptest.NewRequest("GET", "/plans", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
