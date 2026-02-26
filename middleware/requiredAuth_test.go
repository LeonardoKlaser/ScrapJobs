package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testJWTSecret = "test-secret-key-that-is-long-enough"

func createTestToken(userID int, secret string, expired bool) string {
	exp := time.Now().Add(24 * time.Hour)
	if expired {
		exp = time.Now().Add(-1 * time.Hour)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": float64(userID),
		"exp": exp.Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func setupAuthMiddleware(mockUserRepo *mocks.MockUserRepository) (*Middleware, *gin.Engine) {
	userUsecase := usecase.NewUserUsercase(mockUserRepo)
	m := NewMiddleware(userUsecase)

	router := gin.New()
	return m, router
}

func TestRequireAuth_ValidToken(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/protected", m.RequireAuth, func(ctx *gin.Context) {
		user, _ := ctx.Get("user")
		ctx.JSON(http.StatusOK, gin.H{"user": user.(model.User).Name})
	})

	expectedUser := model.User{Id: 1, Name: "Test User", Email: "test@test.com"}
	mockUserRepo.On("GetUserById", 1).Return(expectedUser, nil).Once()

	tokenString := createTestToken(1, testJWTSecret, false)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test User")
	mockUserRepo.AssertExpectations(t)
}

func TestRequireAuth_MissingCookie(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/protected", m.RequireAuth, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidSignature(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/protected", m.RequireAuth, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	tokenString := createTestToken(1, "wrong-secret-key-that-is-different", false)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/protected", m.RequireAuth, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	tokenString := createTestToken(1, testJWTSecret, true)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/protected", m.RequireAuth, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	// Return user with Id=0 (not found)
	mockUserRepo.On("GetUserById", 999).Return(model.User{}, nil).Once()

	tokenString := createTestToken(999, testJWTSecret, false)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUserRepo.AssertExpectations(t)
}
