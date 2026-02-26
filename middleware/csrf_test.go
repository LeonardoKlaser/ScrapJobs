package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCSRFProtection_GETPassesThrough(t *testing.T) {
	router := gin.New()
	router.GET("/data", CSRFProtection(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFProtection_POSTValidOrigin(t *testing.T) {
	router := gin.New()
	router.POST("/data", CSRFProtection(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/data", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFProtection_POSTNoOriginOrReferer(t *testing.T) {
	router := gin.New()
	router.POST("/data", CSRFProtection(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/data", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFProtection_POSTInvalidOrigin(t *testing.T) {
	router := gin.New()
	router.POST("/data", CSRFProtection(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/data", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFProtection_POSTValidRefererFallback(t *testing.T) {
	router := gin.New()
	router.POST("/data", CSRFProtection(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/data", nil)
	req.Header.Set("Referer", "https://scrapjobs.com.br/app/dashboard")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFProtection_POSTFrontendURLEnv(t *testing.T) {
	os.Setenv("FRONTEND_URL", "https://staging.scrapjobs.com")
	defer os.Unsetenv("FRONTEND_URL")

	router := gin.New()
	router.POST("/data", CSRFProtection(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/data", nil)
	req.Header.Set("Origin", "https://staging.scrapjobs.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
