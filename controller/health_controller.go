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
