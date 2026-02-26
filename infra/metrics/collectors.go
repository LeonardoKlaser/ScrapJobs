package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// safeRegister registers collectors, silently ignoring AlreadyRegisteredError.
func safeRegister(cs ...prometheus.Collector) {
	for _, c := range cs {
		if err := prometheus.Register(c); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				panic(err)
			}
		}
	}
}

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

	// GaugeFunc reading a cumulative snapshot from sql.DBStats.WaitCount.
	// Named without _total since it is exposed as a gauge, not a counter.
	dbWaitCount := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_pool_wait_count",
		Help: "Total number of connections waited for (cumulative)",
	}, func() float64 {
		return float64(db.Stats().WaitCount)
	})

	safeRegister(dbOpenConns, dbIdleConns, dbInUseConns, dbWaitCount)
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

	// GaugeFunc reading cumulative snapshots from redis.PoolStats.
	// Named without _total since they are exposed as gauges, not counters.
	redisHits := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "redis_pool_hits",
		Help: "Number of times a free connection was found in the pool (cumulative)",
	}, func() float64 {
		return float64(client.PoolStats().Hits)
	})

	redisMisses := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "redis_pool_misses",
		Help: "Number of times a free connection was NOT found in the pool (cumulative)",
	}, func() float64 {
		return float64(client.PoolStats().Misses)
	})

	safeRegister(redisActiveConns, redisIdleConns, redisHits, redisMisses)
}
