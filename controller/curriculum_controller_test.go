package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupCurriculumController() (*CurriculumController, *mocks.MockCurriculumRepository) {
	mockRepo := new(mocks.MockCurriculumRepository)
	uc := usecase.NewCurriculumUsecase(mockRepo)
	ctrl := NewCurriculumController(uc)
	return ctrl, mockRepo
}

func setUserContext(ctx *gin.Context, user model.User) {
	ctx.Set("user", user)
}

func TestCurriculumController_CreateCurriculum(t *testing.T) {
	ctrl, mockRepo := setupCurriculumController()
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should create curriculum successfully", func(t *testing.T) {
		curriculum := model.Curriculum{Title: "My CV", Skills: "Go, React"}
		expected := model.Curriculum{Id: 1, Title: "My CV", Skills: "Go, React", UserID: 1}

		mockRepo.On("CreateCurriculum", model.Curriculum{Title: "My CV", Skills: "Go, React", UserID: 1}).Return(expected, nil).Once()

		body, _ := json.Marshal(curriculum)
		w := httptest.NewRecorder()
		ctx, router := gin.CreateTestContext(w)

		router.POST("/curriculum", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.CreateCurriculum(c)
		})

		ctx.Request = httptest.NewRequest("POST", "/curriculum", bytes.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, ctx.Request)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid body", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/curriculum", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.CreateCurriculum(c)
		})

		req := httptest.NewRequest("POST", "/curriculum", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCurriculumController_GetCurriculumByUserId(t *testing.T) {
	ctrl, mockRepo := setupCurriculumController()
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should return curricula successfully", func(t *testing.T) {
		expected := []model.Curriculum{
			{Id: 1, Title: "CV 1", UserID: 1},
			{Id: 2, Title: "CV 2", UserID: 1},
		}

		mockRepo.On("FindCurriculumByUserID", 1).Return(expected, nil).Once()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.GET("/curriculum", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.GetCurriculumByUserId(c)
		})

		req := httptest.NewRequest("GET", "/curriculum", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []model.Curriculum
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurriculumController_UpdateCurriculum(t *testing.T) {
	ctrl, mockRepo := setupCurriculumController()
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should update curriculum successfully", func(t *testing.T) {
		curriculum := model.Curriculum{Id: 1, Title: "Updated CV", Skills: "Go"}
		expected := model.Curriculum{Id: 1, Title: "Updated CV", Skills: "Go", UserID: 1}

		mockRepo.On("UpdateCurriculum", model.Curriculum{Id: 1, Title: "Updated CV", Skills: "Go", UserID: 1}).Return(expected, nil).Once()

		body, _ := json.Marshal(curriculum)
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PUT("/curriculum", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.UpdateCurriculum(c)
		})

		req := httptest.NewRequest("PUT", "/curriculum", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurriculumController_DeleteCurriculum(t *testing.T) {
	ctrl, mockRepo := setupCurriculumController()
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should delete curriculum successfully", func(t *testing.T) {
		mockRepo.On("CountCurriculumsByUserID", 1).Return(2, nil).Once()
		mockRepo.On("DeleteCurriculum", 1, 5).Return(nil).Once()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.DELETE("/curriculum/:id", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.DeleteCurriculum(c)
		})

		req := httptest.NewRequest("DELETE", "/curriculum/5", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 when only one curriculum", func(t *testing.T) {
		mockRepo.On("CountCurriculumsByUserID", 1).Return(1, nil).Once()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.DELETE("/curriculum/:id", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.DeleteCurriculum(c)
		})

		req := httptest.NewRequest("DELETE", "/curriculum/5", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.DELETE("/curriculum/:id", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.DeleteCurriculum(c)
		})

		req := httptest.NewRequest("DELETE", "/curriculum/abc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
