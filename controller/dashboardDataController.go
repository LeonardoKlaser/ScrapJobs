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

// GetDashboardDataByUserId godoc
// @Summary Dados do dashboard
// @Description Retorna estatisticas e vagas recentes do usuario
// @Tags Dashboard
// @Produce json
// @Success 200 {object} model.DashboardData
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/dashboard [get]
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

// GetLatestJobs godoc
// @Summary Vagas recentes paginadas
// @Description Retorna vagas com paginacao e filtros
// @Tags Dashboard
// @Produce json
// @Param page query int false "Pagina" default(1)
// @Param limit query int false "Limite por pagina (max 50)" default(10)
// @Param days query int false "Filtrar por dias" default(0)
// @Param search query string false "Buscar por titulo"
// @Success 200 {object} model.PaginatedJobs
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/dashboard/jobs [get]
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