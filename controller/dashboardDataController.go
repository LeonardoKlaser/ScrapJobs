package controller

import (
	"net/http"
	"web-scrapper/model"
	"web-scrapper/repository"

	"github.com/gin-gonic/gin"
)

type DashboardDataController struct{
	repo *repository.DashboardRepository
}

func NewDashboardDataController(rep *repository.DashboardRepository) *DashboardDataController {
	return &DashboardDataController{
		repo: rep,
	}
}

func (repo *DashboardDataController) GetDashboardDataByUserId (ctx *gin.Context) {
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

	data, error := repo.repo.GetDashboardData(user.Id)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}