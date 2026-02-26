# Redis Pooling & Observability Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Centralize Redis client management with proper connection pooling, and add Prometheus metrics for HTTP requests, DB pool, and Redis pool observability.

**Architecture:** Create a shared Redis client in `infra/redis/` injected into all consumers. Add a Prometheus middleware to Gin that tracks request count/duration, plus custom collectors for DB and Redis pool stats. Expose `/metrics` endpoint.

**Tech Stack:** go-redis v9 (already in go.mod), prometheus/client_golang (new dependency), Gin

---

### Task 1: Add prometheus/client_golang dependency

**Files:**
- Modify: `go.mod`

**Step 1: Install the dependency**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go get github.com/prometheus/client_golang/prometheus github.com/prometheus/client_golang/prometheus/promhttp
```

**Step 2: Verify it was added**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && grep prometheus go.mod
```
Expected: line with `github.com/prometheus/client_golang`

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add go.mod go.sum
git commit -m "chore: add prometheus/client_golang dependency"
```

---

### Task 2: Create shared Redis client package

**Files:**
- Create: `infra/redis/client.go`
- Test: `infra/redis/client_test.go`

**Step 1: Create `infra/redis/client.go`**

```go
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a Redis client with a configured connection pool.
// The caller is responsible for calling Close() when done (typically via defer in main).
func NewRedisClient(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     10,
		MinIdleConns: 2,
		PoolTimeout:  5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
```

**Step 2: Create `infra/redis/client_test.go`**

```go
package redis

import (
	"testing"
)

func TestNewRedisClient_InvalidAddr(t *testing.T) {
	_, err := NewRedisClient("localhost:99999")
	if err == nil {
		t.Fatal("expected error for invalid redis address, got nil")
	}
}
```

**Step 3: Run the test**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./infra/redis/ -v
```
Expected: TestNewRedisClient_InvalidAddr PASS (connection refused)

**Step 4: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add infra/redis/
git commit -m "feat: add centralized Redis client with connection pooling"
```

---

### Task 3: Refactor PaymentUsecase to use injected Redis client

**Files:**
- Modify: `usecase/payment_usecase.go`
- Modify: `cmd/api/main.go` (lines 101, 174)
- Modify: `cmd/worker/main.go` (lines 100-103)

**Step 1: Modify `usecase/payment_usecase.go`**

Replace the struct and constructor to accept `redis.UniversalClient` instead of `asynq.RedisConnOpt`:

```go
// payment_usecase.go — new struct
type PaymentUsecase struct {
	paymentGateway *gateway.AbacatePayGateway
	redisClient    redis.UniversalClient
	userUsecase    *UserUsecase
	planRepository interfaces.PlanRepositoryInterface
}

func NewPaymentUsecase(gw *gateway.AbacatePayGateway, redisClient redis.UniversalClient, userUsecase *UserUsecase, planRp interfaces.PlanRepositoryInterface) *PaymentUsecase {
	return &PaymentUsecase{paymentGateway: gw, redisClient: redisClient, userUsecase: userUsecase, planRepository: planRp}
}
```

Remove the `getRedisClient()` method entirely.

In `CreatePayment()`, replace lines 82-83:
```go
// OLD:
redisClient := uc.getRedisClient()
defer redisClient.Close()

// NEW:
redisClient := uc.redisClient
```

In `CompleteRegistration()`, replace lines 100-101:
```go
// OLD:
redisClient := uc.getRedisClient()
defer redisClient.Close()

// NEW:
redisClient := uc.redisClient
```

Remove the `asynq` import (no longer needed in this file). The `redis` import stays.

**Step 2: Modify `cmd/api/main.go`**

Add import for the new redis package:
```go
import (
	// ... existing imports ...
	redispkg "web-scrapper/infra/redis"
)
```

Replace lines 101-102 (`connectRedis` + asynq client creation):
```go
// Create shared Redis client with connection pool
redisClient, err := redispkg.NewRedisClient(secrets.RedisAddr)
if err != nil {
	logging.Logger.Fatal().Err(err).Str("redis_addr", secrets.RedisAddr).Msg("Could not connect to Redis")
}
defer redisClient.Close()
logging.Logger.Info().Str("redis_addr", secrets.RedisAddr).Msg("Connected to Redis")

asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: secrets.RedisAddr})
defer asynqClient.Close()
```

On line 174, change PaymentUsecase construction:
```go
// OLD:
paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, redisOpt, userUsecase, planRepository)

// NEW:
paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, redisClient, userUsecase, planRepository)
```

Remove the `connectRedis()` function (lines 302-312) entirely — no longer needed.

Remove the unused `redis` import from main.go (the `"github.com/redis/go-redis/v9"` one — we now use the wrapper).

**Step 3: Modify `cmd/worker/main.go`**

Add import:
```go
import (
	// ... existing imports ...
	redispkg "web-scrapper/infra/redis"
)
```

Replace lines 100-103 (PaymentUsecase setup with asynq.RedisClientOpt):
```go
// OLD:
redisOpt := asynq.RedisClientOpt{Addr: secrets.RedisAddr}
paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, redisOpt, userUsecase, planRepository)

// NEW:
workerRedisClient, err := redispkg.NewRedisClient(secrets.RedisAddr)
if err != nil {
	logging.Logger.Fatal().Err(err).Msg("Could not connect to Redis for PaymentUsecase")
}
defer workerRedisClient.Close()
paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, workerRedisClient, userUsecase, planRepository)
```

**Step 4: Verify it compiles**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./...
```
Expected: no errors

**Step 5: Run existing tests**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./usecase/... ./cmd/... -v -count=1 2>&1 | tail -30
```
Expected: all tests pass

**Step 6: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add usecase/payment_usecase.go cmd/api/main.go cmd/worker/main.go
git commit -m "refactor: inject shared Redis client into PaymentUsecase

Replace per-call client creation with a shared pooled client.
Eliminates connection churn in CreatePayment and CompleteRegistration."
```

---

### Task 4: Update HealthController to accept Redis client

**Files:**
- Modify: `controller/health_controller.go`
- Modify: `cmd/api/main.go` (line 182)

**Step 1: Modify `controller/health_controller.go`**

Add a `redisClient` field and use it for health check alongside the existing asynq client:

```go
package controller

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type HealthController struct {
	db          *sql.DB
	asynqclient *asynq.Client
	redisClient redis.UniversalClient
}

func NewHealthController(db *sql.DB, asynqClient *asynq.Client, redisClient redis.UniversalClient) *HealthController {
	return &HealthController{
		db:          db,
		asynqclient: asynqClient,
		redisClient: redisClient,
	}
}

func (h *HealthController) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

func (h *HealthController) Readiness(c *gin.Context) {
	dbStatus := "UP"
	if err := h.db.Ping(); err != nil {
		dbStatus = "DOWN"
	}

	redisStatus := "UP"
	if h.redisClient != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			redisStatus = "DOWN"
		}
	} else if err := h.asynqclient.Ping(); err != nil {
		redisStatus = "DOWN"
	}

	if dbStatus == "DOWN" || redisStatus == "DOWN" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"database": dbStatus,
			"redis":    redisStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database": dbStatus,
		"redis":    redisStatus,
	})
}
```

**Step 2: Update `cmd/api/main.go` line 182**

```go
// OLD:
healthController := controller.NewHealthController(dbConnection, asynqClient)

// NEW:
healthController := controller.NewHealthController(dbConnection, asynqClient, redisClient)
```

**Step 3: Verify it compiles and tests pass**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./... && go test ./controller/... -v -count=1 2>&1 | tail -20
```

**Step 4: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add controller/health_controller.go cmd/api/main.go
git commit -m "feat: use pooled Redis client for health checks"
```

---

### Task 5: Create Prometheus metrics middleware

**Files:**
- Create: `infra/metrics/prometheus.go`

**Step 1: Create `infra/metrics/prometheus.go`**

```go
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// GinPrometheus returns a Gin middleware that records request count and duration.
func GinPrometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Use FullPath() to get the route pattern (e.g. "/curriculum/:id")
		// instead of the actual path to avoid high cardinality.
		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
```

**Step 2: Verify it compiles**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./infra/metrics/
```

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add infra/metrics/prometheus.go
git commit -m "feat: add Prometheus middleware for HTTP request metrics"
```

---

### Task 6: Create DB and Redis pool stats collectors

**Files:**
- Create: `infra/metrics/collectors.go`

**Step 1: Create `infra/metrics/collectors.go`**

```go
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
```

**Step 2: Verify it compiles**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./infra/metrics/
```

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add infra/metrics/collectors.go
git commit -m "feat: add DB and Redis pool stats collectors for Prometheus"
```

---

### Task 7: Wire metrics into the API server

**Files:**
- Modify: `cmd/api/main.go`

**Step 1: Add imports**

Add to the import block in `cmd/api/main.go`:
```go
import (
	// ... existing ...
	"web-scrapper/infra/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)
```

**Step 2: Register pool collectors after DB and Redis are created**

After the `redisClient` and `dbConnection` are created (around line ~103), add:
```go
metrics.RegisterDBCollector(dbConnection)
metrics.RegisterRedisCollector(redisClient)
```

**Step 3: Add Prometheus middleware to all route groups**

Add `metrics.GinPrometheus()` to each route group, right after `logging.GinMiddleware()`:

```go
// publicRoutes (around line 207)
publicRoutes.Use(logging.GinMiddleware())
publicRoutes.Use(metrics.GinPrometheus())

// checkoutRoutes (around line 220)
checkoutRoutes.Use(logging.GinMiddleware())
checkoutRoutes.Use(metrics.GinPrometheus())

// privateRoutes (around line 230)
privateRoutes.Use(logging.GinMiddleware())
privateRoutes.Use(metrics.GinPrometheus())

// adminRoutes (around line 262)
adminRoutes.Use(logging.GinMiddleware())
adminRoutes.Use(metrics.GinPrometheus())
```

**Step 4: Add /metrics endpoint**

After the healthRoutes block (around line 276), add:
```go
server.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

**Step 5: Verify it compiles**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./cmd/api/
```

**Step 6: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add cmd/api/main.go
git commit -m "feat: wire Prometheus metrics into API server

- HTTP request count and duration per method/path/status
- DB pool stats (open, idle, in-use, wait count)
- Redis pool stats (active, idle, hits, misses)
- /metrics endpoint for Prometheus scraping"
```

---

### Task 8: Write metrics middleware test

**Files:**
- Create: `infra/metrics/prometheus_test.go`

**Step 1: Create test file**

```go
package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestGinPrometheus_IncrementsCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GinPrometheus())
	router.GET("/test-path", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Reset the counter for this test
	httpRequestsTotal.Reset()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-path", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/test-path", "200"))
	if count != 1 {
		t.Fatalf("expected request count 1, got %f", count)
	}
}

func TestGinPrometheus_RecordsDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GinPrometheus())
	router.GET("/duration-test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/duration-test", nil)
	router.ServeHTTP(w, req)

	// Histogram should have at least 1 observation
	count := testutil.CollectAndCount(httpRequestDuration)
	if count == 0 {
		t.Fatal("expected histogram to have observations")
	}
}

func TestGinPrometheus_UnmatchedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GinPrometheus())
	// No routes registered — any path will be "unmatched"

	httpRequestsTotal.Reset()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "unmatched", "404"))
	if count != 1 {
		t.Fatalf("expected unmatched request count 1, got %f", count)
	}
}
```

**Step 2: Run tests**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./infra/metrics/ -v -count=1
```
Expected: all 3 tests pass

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add infra/metrics/prometheus_test.go
git commit -m "test: add Prometheus middleware tests"
```

---

### Task 9: Final verification

**Step 1: Full build**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./...
```
Expected: all 4 binaries compile

**Step 2: Full test suite**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./... -count=1 2>&1 | tail -40
```
Expected: all tests pass (skip dockertest ones if no Docker)

**Step 3: Race detector**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./infra/metrics/ ./infra/redis/ ./usecase/... ./controller/... ./middleware/... -race -count=1 2>&1 | tail -20
```
Expected: zero data races

**Step 4: Vet**

Run:
```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go vet ./...
```
Expected: no issues
