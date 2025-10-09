package controller

import (
	"fmt"
	"net/http"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type RequestedSiteController struct {
	usecase *usecase.RequestedSiteUsecase
}

func NewRequestedSiteController(usecase *usecase.RequestedSiteUsecase) *RequestedSiteController {
	return &RequestedSiteController{
		usecase: usecase,
	}
}

func (c *RequestedSiteController) Create(ctx *gin.Context) {
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

	var body struct {
		URL string `json:"url"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("corpo da requisição inválido: %w", err).Error()})
		return
	}

	if err := c.usecase.Create(user.Id, body.URL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar a solicitação"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Solicitação enviada com sucesso!"})
}