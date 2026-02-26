package middleware

import (
	"context"
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

func newRateLimiterStore(ctx context.Context) *rateLimiterStore {
	store := &rateLimiterStore{
		visitors: make(map[string]*visitor),
	}
	go store.cleanup(ctx)
	return store
}

func (s *rateLimiterStore) cleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			for ip, v := range s.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(s.visitors, ip)
				}
			}
			s.mu.Unlock()
		}
	}
}

// RateLimiter creates a rate-limiting middleware. Use RateLimiterWithContext to
// allow graceful shutdown of the cleanup goroutine.
func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	return RateLimiterWithContext(context.Background(), r, b)
}

// RateLimiterWithContext creates a rate-limiting middleware whose cleanup
// goroutine stops when ctx is cancelled, preventing goroutine leaks.
func RateLimiterWithContext(ctx context.Context, r rate.Limit, b int) gin.HandlerFunc {
	store := newRateLimiterStore(ctx)

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
