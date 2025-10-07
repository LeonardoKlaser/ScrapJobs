package controller

import (
	"net/http"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type PlanController struct {
	usecase *usecase.PlanUsecase
}

func NewPlanController(usecase *usecase.PlanUsecase) *PlanController {
	return &PlanController{
		usecase: usecase,
	}
}

func (uc *PlanController) GetAllPlans(ctx *gin.Context) {
	plans, err := uc.usecase.GetAllPlans()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar planos"})
		return
	}

	ctx.JSON(http.StatusOK, plans)
}