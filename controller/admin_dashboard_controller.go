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

// GetAdminDashboard godoc
// @Summary Dashboard administrativo
// @Description Retorna metricas administrativas (somente admin)
// @Tags Admin
// @Produce json
// @Success 200 {object} model.AdminDashboardData
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/admin/dashboard [get]
func (c *AdminDashboardController) GetAdminDashboard(ctx *gin.Context) {
	data, err := c.repo.GetAdminDashboardData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
