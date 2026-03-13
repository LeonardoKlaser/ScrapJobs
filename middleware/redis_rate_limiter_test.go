package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestRedisClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestRedisRateLimiter_WithinLimit(t *testing.T) {
	client, _ := newTestRedisClient(t)
	defer client.Close()

	router := gin.New()
	router.GET("/api", RedisRateLimiter(client, "test", 5, 60), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should pass", i+1)
	}
}

func TestRedisRateLimiter_ExceedsLimit(t *testing.T) {
	client, _ := newTestRedisClient(t)
	defer client.Close()

	router := gin.New()
	router.GET("/api", RedisRateLimiter(client, "test", 2, 60), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRedisRateLimiter_DifferentIPs(t *testing.T) {
	client, _ := newTestRedisClient(t)
	defer client.Close()

	router := gin.New()
	router.GET("/api", RedisRateLimiter(client, "test", 1, 60), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// IP A: 1 request — should pass
	reqA := httptest.NewRequest("GET", "/api", nil)
	reqA.RemoteAddr = "10.0.0.1:12345"
	wA := httptest.NewRecorder()
	router.ServeHTTP(wA, reqA)
	assert.Equal(t, http.StatusOK, wA.Code)

	// IP A: 2nd request — should be limited
	reqA2 := httptest.NewRequest("GET", "/api", nil)
	reqA2.RemoteAddr = "10.0.0.1:12345"
	wA2 := httptest.NewRecorder()
	router.ServeHTTP(wA2, reqA2)
	assert.Equal(t, http.StatusTooManyRequests, wA2.Code)

	// IP B: 1st request — should pass (independent counter)
	reqB := httptest.NewRequest("GET", "/api", nil)
	reqB.RemoteAddr = "10.0.0.2:12345"
	wB := httptest.NewRecorder()
	router.ServeHTTP(wB, reqB)
	assert.Equal(t, http.StatusOK, wB.Code)
}

func TestRedisRateLimiter_WindowExpires(t *testing.T) {
	client, mr := newTestRedisClient(t)
	defer client.Close()

	router := gin.New()
	router.GET("/api", RedisRateLimiter(client, "test", 1, 60), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 1st request passes
	req := httptest.NewRequest("GET", "/api", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2nd request blocked
	req2 := httptest.NewRequest("GET", "/api", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Fast-forward time so the window expires
	mr.FastForward(61 * 1e9) // 61 seconds in nanoseconds

	// 3rd request should pass again
	req3 := httptest.NewRequest("GET", "/api", nil)
	req3.RemoteAddr = "10.0.0.1:12345"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRedisRateLimiter_SharedCounterAcrossInstances(t *testing.T) {
	client, _ := newTestRedisClient(t)
	defer client.Close()

	// Two middleware instances sharing the same Redis client (simulates 2 API instances)
	router1 := gin.New()
	router1.GET("/api", RedisRateLimiter(client, "test", 2, 60), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router2 := gin.New()
	router2.GET("/api", RedisRateLimiter(client, "test", 2, 60), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Instance 1: 1st request — passes (count=1)
	req1 := httptest.NewRequest("GET", "/api", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	w1 := httptest.NewRecorder()
	router1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Instance 2: 2nd request — passes (count=2)
	req2 := httptest.NewRequest("GET", "/api", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Instance 1: 3rd request — blocked (count=3 > limit=2)
	req3 := httptest.NewRequest("GET", "/api", nil)
	req3.RemoteAddr = "10.0.0.1:12345"
	w3 := httptest.NewRecorder()
	router1.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}
