package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// luaRateLimit is an atomic fixed-window counter script.
// It increments the counter for a key, sets TTL on first hit,
// and returns 0 if the limit is exceeded, 1 otherwise.
var luaRateLimit = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end
if current > limit then
    return 0
end
return 1
`)

// RedisRateLimiter creates a distributed rate-limiting middleware backed by Redis.
// It uses a fixed-window counter with an atomic Lua script.
// limiterName namespaces the Redis key so different limiters don't share counters.
// limit is the max number of requests per window, windowSeconds is the window duration.
func RedisRateLimiter(redisClient *redis.Client, limiterName string, limit int, windowSeconds int) gin.HandlerFunc {
	limitStr := strconv.Itoa(limit)
	windowStr := strconv.Itoa(windowSeconds)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate_limit:" + limiterName + ":" + ip

		result, err := luaRateLimit.Run(c.Request.Context(), redisClient, []string{key}, limitStr, windowStr).Int()
		if err != nil {
			// If Redis is unavailable, allow the request through (fail open).
			c.Next()
			return
		}

		if result == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Muitas requisições. Tente novamente em alguns segundos"})
			return
		}

		c.Next()
	}
}
