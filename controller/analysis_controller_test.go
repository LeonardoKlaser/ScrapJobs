package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAnalysisController() (*AnalysisController, *mocks.MockAnalysisService, *mocks.MockCurriculumRepository, *mocks.MockJobRepository, *mocks.MockNotificationRepository, *mocks.MockPlanRepository, *mocks.MockEmailService) {
	mockAnalysis := new(mocks.MockAnalysisService)
	mockCurriculum := new(mocks.MockCurriculumRepository)
	mockJob := new(mocks.MockJobRepository)
	mockNotification := new(mocks.MockNotificationRepository)
	mockPlan := new(mocks.MockPlanRepository)
	mockEmail := new(mocks.MockEmailService)

	ctrl := NewAnalysisController(mockAnalysis, mockCurriculum, mockJob, mockNotification, mockPlan, mockEmail)
	return ctrl, mockAnalysis, mockCurriculum, mockJob, mockNotification, mockPlan, mockEmail
}

func TestAnalysisController_AnalyzeJob(t *testing.T) {
	t.Run("should return 401 when user not in context", func(t *testing.T) {
		ctrl, _, _, _, _, _, _ := setupAnalysisController()

		body, _ := json.Marshal(map[string]int{"job_id": 1})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", ctrl.AnalyzeJob)

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 when job_id is missing", func(t *testing.T) {
		ctrl, _, _, _, _, _, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		body, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "job_id")
	})

	t.Run("should return 400 when no curriculum found", func(t *testing.T) {
		ctrl, _, mockCurriculum, _, _, _, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		// Return empty curricula list
		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{}, nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "currículo")
		mockCurriculum.AssertExpectations(t)
	})

	t.Run("should return 403 when no plan associated", func(t *testing.T) {
		ctrl, _, mockCurriculum, _, _, mockPlan, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{
			{Id: 1, Title: "CV 1", UserID: 1},
		}, nil).Once()
		mockPlan.On("GetPlanByUserID", 1).Return(nil, nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockCurriculum.AssertExpectations(t)
		mockPlan.AssertExpectations(t)
	})

	t.Run("should return 403 when monthly analysis limit reached", func(t *testing.T) {
		ctrl, _, mockCurriculum, _, mockNotification, mockPlan, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{
			{Id: 1, Title: "CV 1", UserID: 1},
		}, nil).Once()
		mockPlan.On("GetPlanByUserID", 1).Return(&model.Plan{ID: 1, MaxAIAnalyses: 10}, nil).Once()
		mockNotification.On("GetMonthlyAnalysisCount", 1).Return(10, nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "Limite de análises")
		mockCurriculum.AssertExpectations(t)
		mockPlan.AssertExpectations(t)
		mockNotification.AssertExpectations(t)
	})

	t.Run("should skip quota check when MaxAIAnalyses is zero", func(t *testing.T) {
		ctrl, mockAnalysis, mockCurriculum, mockJob, mockNotification, mockPlan, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		curriculum := model.Curriculum{Id: 1, Title: "CV 1", UserID: 1, Skills: "Go"}
		job := &model.Job{ID: 10, Title: "Go Dev", Company: "TestCo"}
		analysis := model.ResumeAnalysis{
			MatchAnalysis:       model.MatchAnalysis{OverallScoreNumeric: 80},
			FinalConsiderations: "OK",
		}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{curriculum}, nil).Once()
		// MaxAIAnalyses = 0 means the guard (> 0) skips quota check entirely
		mockPlan.On("GetPlanByUserID", 1).Return(&model.Plan{ID: 1, MaxAIAnalyses: 0}, nil).Once()
		// GetMonthlyAnalysisCount should NOT be called because guard skips it
		mockJob.On("GetJobByID", 10).Return(job, nil).Once()
		mockAnalysis.On("Analyze", mock.Anything, curriculum, *job).Return(analysis, nil).Once()
		mockNotification.On("InsertNotificationWithAnalysis", 10, 1, 1, mock.Anything).Return(nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10, "curriculum_id": 1})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockNotification.AssertNotCalled(t, "GetMonthlyAnalysisCount", mock.Anything)
		mockCurriculum.AssertExpectations(t)
		mockPlan.AssertExpectations(t)
		mockJob.AssertExpectations(t)
		mockAnalysis.AssertExpectations(t)
	})

	t.Run("should skip quota check when MaxAIAnalyses is negative (legacy unlimited)", func(t *testing.T) {
		ctrl, mockAnalysis, mockCurriculum, mockJob, mockNotification, mockPlan, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		curriculum := model.Curriculum{Id: 1, Title: "CV 1", UserID: 1, Skills: "Go"}
		job := &model.Job{ID: 10, Title: "Go Dev", Company: "TestCo"}
		analysis := model.ResumeAnalysis{
			MatchAnalysis:       model.MatchAnalysis{OverallScoreNumeric: 90},
			FinalConsiderations: "Great",
		}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{curriculum}, nil).Once()
		// MaxAIAnalyses = -1 (legacy unlimited) — guard (> 0) skips quota check
		mockPlan.On("GetPlanByUserID", 1).Return(&model.Plan{ID: 1, MaxAIAnalyses: -1}, nil).Once()
		mockJob.On("GetJobByID", 10).Return(job, nil).Once()
		mockAnalysis.On("Analyze", mock.Anything, curriculum, *job).Return(analysis, nil).Once()
		mockNotification.On("InsertNotificationWithAnalysis", 10, 1, 1, mock.Anything).Return(nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10, "curriculum_id": 1})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockNotification.AssertNotCalled(t, "GetMonthlyAnalysisCount", mock.Anything)
		mockCurriculum.AssertExpectations(t)
		mockPlan.AssertExpectations(t)
		mockJob.AssertExpectations(t)
		mockAnalysis.AssertExpectations(t)
	})

	t.Run("should return 404 when job not found", func(t *testing.T) {
		ctrl, _, mockCurriculum, mockJob, mockNotification, mockPlan, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{
			{Id: 1, Title: "CV 1", UserID: 1},
		}, nil).Once()
		mockPlan.On("GetPlanByUserID", 1).Return(&model.Plan{ID: 1, MaxAIAnalyses: 10}, nil).Once()
		mockNotification.On("GetMonthlyAnalysisCount", 1).Return(5, nil).Once()
		mockJob.On("GetJobByID", 999).Return(nil, nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 999})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockCurriculum.AssertExpectations(t)
		mockPlan.AssertExpectations(t)
		mockNotification.AssertExpectations(t)
		mockJob.AssertExpectations(t)
	})

	t.Run("should return 200 with analysis on success", func(t *testing.T) {
		ctrl, mockAnalysis, mockCurriculum, mockJob, mockNotification, mockPlan, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		curriculum := model.Curriculum{Id: 1, Title: "CV 1", UserID: 1, Skills: "Go, React"}
		job := &model.Job{ID: 10, Title: "Go Dev", Company: "TestCo"}
		analysis := model.ResumeAnalysis{
			MatchAnalysis:       model.MatchAnalysis{OverallScoreNumeric: 85},
			FinalConsiderations: "Good match",
		}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum{curriculum}, nil).Once()
		mockPlan.On("GetPlanByUserID", 1).Return(&model.Plan{ID: 1, MaxAIAnalyses: 100}, nil).Once()
		mockNotification.On("GetMonthlyAnalysisCount", 1).Return(5, nil).Once()
		mockJob.On("GetJobByID", 10).Return(job, nil).Once()
		mockAnalysis.On("Analyze", mock.Anything, curriculum, *job).Return(analysis, nil).Once()
		mockNotification.On("InsertNotificationWithAnalysis", 10, 1, 1, mock.Anything).Return(nil).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10, "curriculum_id": 1})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp model.ResumeAnalysis
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 85, resp.MatchAnalysis.OverallScoreNumeric)
		mockCurriculum.AssertExpectations(t)
		mockPlan.AssertExpectations(t)
		mockNotification.AssertExpectations(t)
		mockJob.AssertExpectations(t)
		mockAnalysis.AssertExpectations(t)
	})

	t.Run("should return 500 when curriculum fetch fails", func(t *testing.T) {
		ctrl, _, mockCurriculum, _, _, _, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		mockCurriculum.On("FindCurriculumByUserID", 1).Return([]model.Curriculum(nil), errors.New("db error")).Once()

		body, _ := json.Marshal(map[string]int{"job_id": 10})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/analyze", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.AnalyzeJob(c)
		})

		req := httptest.NewRequest("POST", "/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockCurriculum.AssertExpectations(t)
	})
}

func TestAnalysisController_SendAnalysisEmail(t *testing.T) {
	t.Run("should return 401 when user not in context", func(t *testing.T) {
		ctrl, _, _, _, _, _, _ := setupAnalysisController()

		body, _ := json.Marshal(map[string]interface{}{"job_id": 1, "analysis": map[string]string{}})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/send-email", ctrl.SendAnalysisEmail)

		req := httptest.NewRequest("POST", "/send-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for missing job_id", func(t *testing.T) {
		ctrl, _, _, _, _, _, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		body, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/send-email", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.SendAnalysisEmail(c)
		})

		req := httptest.NewRequest("POST", "/send-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 when job not found", func(t *testing.T) {
		ctrl, _, _, mockJob, _, _, _ := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		mockJob.On("GetJobByID", 999).Return(nil, nil).Once()

		analysis := model.ResumeAnalysis{FinalConsiderations: "test"}
		body, _ := json.Marshal(sendAnalysisEmailRequest{JobID: 999, Analysis: analysis})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/send-email", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.SendAnalysisEmail(c)
		})

		req := httptest.NewRequest("POST", "/send-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockJob.AssertExpectations(t)
	})

	t.Run("should return 200 on successful email send", func(t *testing.T) {
		ctrl, _, _, mockJob, _, _, mockEmail := setupAnalysisController()
		user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

		job := &model.Job{ID: 10, Title: "Go Dev", Company: "TestCo"}
		analysis := model.ResumeAnalysis{FinalConsiderations: "Good match"}

		mockJob.On("GetJobByID", 10).Return(job, nil).Once()
		mockEmail.On("SendAnalysisEmail", mock.Anything, "test@test.com", *job, analysis).Return(nil).Once()

		body, _ := json.Marshal(sendAnalysisEmailRequest{JobID: 10, Analysis: analysis})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/send-email", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.SendAnalysisEmail(c)
		})

		req := httptest.NewRequest("POST", "/send-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Email enviado com sucesso", resp["message"])
		mockJob.AssertExpectations(t)
		mockEmail.AssertExpectations(t)
	})
}
