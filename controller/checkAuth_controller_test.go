package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testJWTSecret = "test-secret-key-that-is-long-enough-32chars"

func createTestJWT(claims jwt.MapClaims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func validClaims(userID int) jwt.MapClaims {
	return jwt.MapClaims{
		"sub":       float64(userID),
		"email":     "test@test.com",
		"user_name": "Test User",
		"is_admin":  false,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}
}

func setupCheckAuthTest() (*CheckAuthController, *mocks.MockUserRepository) {
	mockUserRepo := new(mocks.MockUserRepository)
	ctrl := NewCheckAuthController(mockUserRepo)
	return &ctrl, mockUserRepo
}

func TestCheckAuthUser_ValidJWT(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	ctrl, mockUserRepo := setupCheckAuthTest()

	plan := model.Plan{ID: 1, Name: "Profissional", MaxSites: 8, MaxAIAnalyses: 40}
	meData := model.UserMeData{
		UserName:             "Test User",
		IsAdmin:              true, // DB says admin — JWT claim says false → DB must win
		Tax:                  nil,
		WeekdaysOnly:         false,
		Plan:                 &plan,
		MonitoredSitesCount:  3,
		MonthlyAnalysisCount: 5,
	}
	mockUserRepo.On("GetUserMeData", 1).Return(meData, nil).Once()

	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	_ = ctx

	router.GET("/me", func(c *gin.Context) {
		c.Set("claims", validClaims(1)) // JWT has is_admin: false
		ctrl.CheckAuthUser(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_name":"Test User"`)
	assert.Contains(t, w.Body.String(), `"email":"test@test.com"`)
	assert.Contains(t, w.Body.String(), `"is_admin":true`) // DB value, not JWT
	assert.Contains(t, w.Body.String(), `"monthly_analysis_count":5`)
	mockUserRepo.AssertExpectations(t)
}

func TestCheckAuthUser_DeletedUser_Returns401(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	ctrl, mockUserRepo := setupCheckAuthTest()

	mockUserRepo.On("GetUserMeData", 1).Return(
		model.UserMeData{},
		model.ErrUserNotFound,
	).Once()

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	router.GET("/me", func(c *gin.Context) {
		c.Set("claims", validClaims(1))
		ctrl.CheckAuthUser(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Conta desativada")
	mockUserRepo.AssertExpectations(t)
}

func TestCheckAuthUser_DBError_Returns500(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	ctrl, mockUserRepo := setupCheckAuthTest()

	mockUserRepo.On("GetUserMeData", 1).Return(
		model.UserMeData{},
		errors.New("connection refused"),
	).Once()

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	router.GET("/me", func(c *gin.Context) {
		c.Set("claims", validClaims(1))
		ctrl.CheckAuthUser(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUserRepo.AssertExpectations(t)
}

func TestCheckAuthUser_NoClaims_Returns401(t *testing.T) {
	ctrl, _ := setupCheckAuthTest()

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	router.GET("/me", ctrl.CheckAuthUser)

	req := httptest.NewRequest("GET", "/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCheckAuthUser_MalformedSub_Returns401(t *testing.T) {
	ctrl, _ := setupCheckAuthTest()

	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	router.GET("/me", func(c *gin.Context) {
		claims := jwt.MapClaims{
			"sub":   true, // invalid type for sub
			"email": "test@test.com",
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		}
		c.Set("claims", claims)
		ctrl.CheckAuthUser(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token inválido")
}
