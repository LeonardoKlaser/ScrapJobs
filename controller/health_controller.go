package controller

import(
	"database/sql"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type HealthController struct {
	db		*sql.DB
	asynqclient *asynq.Client
}

func NewHealthController(db *sql.DB, asynqClient *asynq.Client) *HealthController{
	return &HealthController{
		db: db,
		asynqclient: asynqClient,
	}
}

func (h *HealthController) Liveness(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{
		"status" : "UP",
	})
}

func (h *HealthController) Readiness(c *gin.Context){
	dbStatus := "UP"
	if err := h.db.Ping(); err != nil{
		dbStatus = "DOWN"
	}

	redisStatus := "UP"
	if err := h.asynqclient.Ping(); err != nil {
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