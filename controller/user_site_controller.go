package controller

import (
	"fmt"
	"net/http"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)


type UserSiteController struct{
	usecase *usecase.UserSiteUsecase
}

func NewUserSiteController(usecase *usecase.UserSiteUsecase) *UserSiteController{
	return &UserSiteController{
		usecase: usecase,
	}
}

func (usecase *UserSiteController) InsertUserSite(ctx *gin.Context) {
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

	var body model.UserSiteRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error to deserialize new job json body: %w", err)})
		return
	}

	err := usecase.usecase.InsertUserSite(user.Id, body.SiteId, body.TargetWords)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error registering user on the website: %w",err)})
		return
	}


	ctx.JSON(http.StatusCreated, gin.H{})
}
