package controller

import (
	"net/http"
	"os"
	"web-scrapper/model"
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
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido no contexto"})
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if !user.IsAdmin && user.Email != adminEmail {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	data, err := c.repo.GetAdminDashboardData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
