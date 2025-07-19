package controller

import (
	"web-scrapper/model"
	"web-scrapper/usecase"
	"net/http"
	"fmt"
	"github.com/gin-gonic/gin"
)

type SiteCareerController struct{
	usecase *usecase.SiteCareerUsecase
}

func NewSiteCareerController(usecase *usecase.SiteCareerUsecase) *SiteCareerController{
	return &SiteCareerController{
		usecase: usecase,
	}
}

func (usecase *SiteCareerController) InsertNewSiteCareer(ctx *gin.Context){
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

	if user.Email != "adminScrapjobs@gmail.com"{
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "only admins can add new sites"})
	}

	var body model.SiteScrapingConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error to deserialize new job json body: %w", err)})
	}

	res, err := usecase.usecase.InsertNewSiteCareer(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error" : fmt.Errorf("ERROR to insert new site career:  %w", err),
		})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}