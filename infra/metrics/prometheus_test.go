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

	count := testutil.CollectAndCount(httpRequestDuration)
	if count == 0 {
		t.Fatal("expected histogram to have observations")
	}
}

func TestGinPrometheus_UnmatchedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GinPrometheus())

	httpRequestsTotal.Reset()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "unmatched", "404"))
	if count != 1 {
		t.Fatalf("expected unmatched request count 1, got %f", count)
	}
}
