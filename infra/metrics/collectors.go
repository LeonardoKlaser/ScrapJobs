package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// RegisterDBCollector registers gauges that track sql.DB pool stats.
func RegisterDBCollector(db *sql.DB) {
	dbOpenConns := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_open_connections",
		Help: "Number of open connections to the database",
	}, func() float64 {
		return float64(db.Stats().OpenConnections)
	})

	dbIdleConns := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_idle_connections",
		Help: "Number of idle connections in the database pool",
	}, func() float64 {
		return float64(db.Stats().Idle)
	})

	dbInUseConns := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_in_use_connections",
		Help: "Number of connections currently in use",
	}, func() float64 {
		return float64(db.Stats().InUse)
	})

	dbWaitCount := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_wait_count_total",
		Help: "Total number of connections waited for",
	}, func() float64 {
		return float64(db.Stats().WaitCount)
	})

	prometheus.MustRegister(dbOpenConns, dbIdleConns, dbInUseConns, dbWaitCount)
}

// RegisterRedisCollector registers gauges that track go-redis pool stats.
func RegisterRedisCollector(client *redis.Client) {
	redisActiveConns := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "redis_pool_active_connections",
		Help: "Number of active Redis connections",
	}, func() float64 {
		return float64(client.PoolStats().TotalConns)
	})

	redisIdleConns := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "redis_pool_idle_connections",
		Help: "Number of idle Redis connections in the pool",
	}, func() float64 {
		return float64(client.PoolStats().IdleConns)
	})

	redisHits := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "redis_pool_hits_total",
		Help: "Number of times a free connection was found in the pool",
	}, func() float64 {
		return float64(client.PoolStats().Hits)
	})

	redisMisses := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "redis_pool_misses_total",
		Help: "Number of times a free connection was NOT found in the pool",
	}, func() float64 {
		return float64(client.PoolStats().Misses)
	})

	prometheus.MustRegister(redisActiveConns, redisIdleConns, redisHits, redisMisses)
}
