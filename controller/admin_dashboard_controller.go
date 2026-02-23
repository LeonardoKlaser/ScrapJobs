package controller

import (
	"net/http"
	"web-scrapper/repository"

	"github.com/gin-gonic/gin"
)

type AdminDashboardController struct {
	repo *repository.DashboardRepository
}

func NewAdminDashboardController(rep *repository.DashboardRepository) *AdminDashboardController {
	return &AdminDashboardController{
		repo: rep,
	}
}

func (c *AdminDashboardController) GetAdminDashboard(ctx *gin.Context) {
	data, err := c.repo.GetAdminDashboardData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
