package controller

import (
	"encoding/json"
	"net/http"
	"time"
	"web-scrapper/logging"
	"web-scrapper/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type StatsController struct {
	dashboardRepo *repository.DashboardRepository
	redisClient   *redis.Client
}

func NewStatsController(dr *repository.DashboardRepository, rc *redis.Client) *StatsController {
	return &StatsController{dashboardRepo: dr, redisClient: rc}
}

func (sc *StatsController) GetPublicStats(ctx *gin.Context) {
	const cacheKey = "public:stats"
	const cacheTTL = 5 * time.Minute

	// Try cache first
	if sc.redisClient != nil {
		cached, err := sc.redisClient.Get(ctx.Request.Context(), cacheKey).Result()
		if err == nil {
			var stats repository.PublicStats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				ctx.JSON(http.StatusOK, stats)
				return
			}
		}
	}

	stats, err := sc.dashboardRepo.GetPublicStats()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Failed to get public stats")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar estatísticas"})
		return
	}

	// Cache result
	if sc.redisClient != nil {
		if data, err := json.Marshal(stats); err == nil {
			sc.redisClient.Set(ctx.Request.Context(), cacheKey, data, cacheTTL)
		}
	}

	ctx.JSON(http.StatusOK, stats)
}
