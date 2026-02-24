package controller

import (
	"net/http"
	"strconv"
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

func (repo *DashboardDataController) GetLatestJobs(ctx *gin.Context) {
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

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "0"))
	search := ctx.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	data, err := repo.repo.GetLatestJobsPaginated(user.Id, page, limit, days, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}