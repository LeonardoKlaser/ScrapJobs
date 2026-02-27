package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a Redis client with a configured connection pool.
// addr can be a plain host:port or a redis:// URL (e.g. from Railway's REDIS_URL).
// The caller is responsible for calling Close() when done (typically via defer in main).
func NewRedisClient(addr string) (*redis.Client, error) {
	var opt *redis.Options
	// Try parsing as URL first (Railway provides REDIS_URL in redis://... format)
	parsed, err := redis.ParseURL(addr)
	if err != nil {
		// Fallback: treat as plain host:port
		opt = &redis.Options{Addr: addr}
	} else {
		opt = parsed
	}
	opt.PoolSize = 10
	opt.MinIdleConns = 2
	opt.PoolTimeout = 5 * time.Second

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
