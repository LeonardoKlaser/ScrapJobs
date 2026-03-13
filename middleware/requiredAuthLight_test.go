package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"web-scrapper/repository/mocks"
)

func TestRequireAuthLight_ValidToken(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	var capturedClaims jwt.MapClaims
	router.GET("/me", m.RequireAuthLight, func(ctx *gin.Context) {
		claimsInterface, _ := ctx.Get("claims")
		capturedClaims = claimsInterface.(jwt.MapClaims)
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tokenString := createTestToken(42, testJWTSecret, false)

	req := httptest.NewRequest("GET", "/me", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, float64(42), capturedClaims["sub"])
	// Verify no DB call was made (RequireAuthLight should NOT call GetUserById)
	mockUserRepo.AssertNotCalled(t, "GetUserById")
}

func TestRequireAuthLight_MissingCookie(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/me", m.RequireAuthLight, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuthLight_ExpiredToken(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/me", m.RequireAuthLight, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	tokenString := createTestToken(1, testJWTSecret, true)

	req := httptest.NewRequest("GET", "/me", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuthLight_InvalidSignature(t *testing.T) {
	os.Setenv("JWTTOKEN", testJWTSecret)
	defer os.Unsetenv("JWTTOKEN")

	mockUserRepo := new(mocks.MockUserRepository)
	m, router := setupAuthMiddleware(mockUserRepo)

	router.GET("/me", m.RequireAuthLight, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})

	tokenString := createTestToken(1, "wrong-secret-totally-different-key", false)

	req := httptest.NewRequest("GET", "/me", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: tokenString})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
