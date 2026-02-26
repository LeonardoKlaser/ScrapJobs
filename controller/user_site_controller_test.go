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

func setupUserSiteController() (*UserSiteController, *mocks.MockUserSiteRepository, *mocks.MockPlanRepository) {
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockPlanRepo := new(mocks.MockPlanRepository)
	uc := usecase.NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)
	ctrl := NewUserSiteController(uc)
	return ctrl, mockUserSiteRepo, mockPlanRepo
}

func TestUserSiteController_InsertUserSite(t *testing.T) {
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should insert user site successfully", func(t *testing.T) {
		ctrl, mockUserSiteRepo, mockPlanRepo := setupUserSiteController()

		plan := &model.Plan{ID: 1, MaxSites: 10}
		mockPlanRepo.On("GetPlanByUserID", 1).Return(plan, nil).Once()
		mockUserSiteRepo.On("GetUserSiteCount", 1).Return(2, nil).Once()
		mockUserSiteRepo.On("InsertNewUserSite", 1, 5, []string{"golang"}).Return(nil).Once()

		body, _ := json.Marshal(model.UserSiteRequest{SiteId: 5, TargetWords: []string{"golang"}})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/userSite", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.InsertUserSite(c)
		})

		req := httptest.NewRequest("POST", "/userSite", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockPlanRepo.AssertExpectations(t)
		mockUserSiteRepo.AssertExpectations(t)
	})

	t.Run("should return 400 when site limit exceeded", func(t *testing.T) {
		ctrl, mockUserSiteRepo, mockPlanRepo := setupUserSiteController()

		plan := &model.Plan{ID: 1, MaxSites: 3}
		mockPlanRepo.On("GetPlanByUserID", 1).Return(plan, nil).Once()
		mockUserSiteRepo.On("GetUserSiteCount", 1).Return(3, nil).Once()

		body, _ := json.Marshal(model.UserSiteRequest{SiteId: 5, TargetWords: []string{"golang"}})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.POST("/userSite", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.InsertUserSite(c)
		})

		req := httptest.NewRequest("POST", "/userSite", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockPlanRepo.AssertExpectations(t)
		mockUserSiteRepo.AssertExpectations(t)
	})
}

func TestUserSiteController_DeleteUserSite(t *testing.T) {
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should delete user site successfully", func(t *testing.T) {
		ctrl, mockUserSiteRepo, _ := setupUserSiteController()

		mockUserSiteRepo.On("DeleteUserSite", 1, "10").Return(nil).Once()

		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.DELETE("/userSite/:siteId", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.DeleteUserSite(c)
		})

		req := httptest.NewRequest("DELETE", "/userSite/10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUserSiteRepo.AssertExpectations(t)
	})
}

func TestUserSiteController_UpdateUserSiteFilters(t *testing.T) {
	user := model.User{Id: 1, Name: "Test", Email: "test@test.com"}

	t.Run("should update filters successfully", func(t *testing.T) {
		ctrl, mockUserSiteRepo, _ := setupUserSiteController()

		mockUserSiteRepo.On("UpdateUserSiteFilters", 1, 10, []string{"go", "backend"}).Return(nil).Once()

		body, _ := json.Marshal(map[string]interface{}{"target_words": []string{"go", "backend"}})
		w := httptest.NewRecorder()
		_, router := gin.CreateTestContext(w)

		router.PATCH("/userSite/:siteId", func(c *gin.Context) {
			setUserContext(c, user)
			ctrl.UpdateUserSiteFilters(c)
		})

		req := httptest.NewRequest("PATCH", "/userSite/10", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUserSiteRepo.AssertExpectations(t)
	})
}
