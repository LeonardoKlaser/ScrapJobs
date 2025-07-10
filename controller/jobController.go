package controller

import (
	"context"
	"net/http"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type jobController struct {
	orchestrator *usecase.TaskEnqueuer
}

func NewJobController(orchestrator *usecase.TaskEnqueuer) jobController {
	return jobController{
		orchestrator: orchestrator,
	}
}

func (c *jobController) ScrappeAndInsert(ctx *gin.Context){
	
	go c.orchestrator.ExecuteScrapingCycle(context.Background())

	ctx.JSON(http.StatusCreated, gin.H{"Message" : "Scrapping Cycle initialized in background"})
}