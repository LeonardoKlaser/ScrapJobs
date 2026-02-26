package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter_FirstRequest(t *testing.T) {
	router := gin.New()
	router.GET("/api", RateLimiter(rate.Limit(10), 5), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_WithinLimit(t *testing.T) {
	router := gin.New()
	router.GET("/api", RateLimiter(rate.Limit(10), 5), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request creates the visitor (always passes), then 5 more within burst limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "request %d should pass", i+1)
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	router := gin.New()
	// Very restrictive: 0.001 requests/sec, burst of 1
	router.GET("/api", RateLimiter(rate.Limit(0.001), 1), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request: creates visitor, always passes
	req1 := httptest.NewRequest("GET", "/api", nil)
	req1.RemoteAddr = "172.16.0.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request: uses the token (burst=1), should pass
	req2 := httptest.NewRequest("GET", "/api", nil)
	req2.RemoteAddr = "172.16.0.1:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request: no tokens left, should be rate limited
	req3 := httptest.NewRequest("GET", "/api", nil)
	req3.RemoteAddr = "172.16.0.1:12345"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}
