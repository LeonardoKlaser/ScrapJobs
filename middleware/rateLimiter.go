package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

func newRateLimiterStore() *rateLimiterStore {
	store := &rateLimiterStore{
		visitors: make(map[string]*visitor),
	}
	go store.cleanup()
	return store
}

func (s *rateLimiterStore) cleanup() {
	for {
		time.Sleep(time.Minute)

		s.mu.Lock()
		for ip, v := range s.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(s.visitors, ip)
			}
		}
		s.mu.Unlock()
	}
}

func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	store := newRateLimiterStore()

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		store.mu.Lock()

		v, exists := store.visitors[ip]
		if !exists {
			limiter := rate.NewLimiter(r, b)
			store.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
			store.mu.Unlock()
			ctx.Next()
			return
		}

		v.lastSeen = time.Now()
		store.mu.Unlock()

		if !v.limiter.Allow() {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		ctx.Next()
	}
}
