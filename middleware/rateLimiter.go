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

var (
    visitors = make(map[string]*visitor)
    mu       sync.Mutex
)

func init() {
    go cleanupVisitors()
}


func RateLimiter(r rate.Limit, b int) gin.HandlerFunc{
	return func(ctx *gin.Context){
		ip := ctx.ClientIP()

		mu.Lock()

		v, exists := visitors[ip]
		if !exists {
			limiter := rate.NewLimiter(r, b)
			visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
			mu.Unlock()
			ctx.Next()
			return
		}

		v.lastSeen = time.Now()
		mu.Unlock()

		if !v.limiter.Allow(){
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
            return
		}

		ctx.Next()
	}
}


func cleanupVisitors() {
    for {
        time.Sleep(time.Minute) 

        mu.Lock()
        for ip, v := range visitors {
            if time.Since(v.lastSeen) > 3*time.Minute {
                delete(visitors, ip)
            }
        }
        mu.Unlock()
    }
}