package controller

import (
	"net/http"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

type CheckAuthController struct {
	
}

func NewCheckAuthController () CheckAuthController {
	return CheckAuthController{
	}
}

func (controller *CheckAuthController) CheckAuthUser( ctx *gin.Context){
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

	ctx.JSON(http.StatusOK, user)
}