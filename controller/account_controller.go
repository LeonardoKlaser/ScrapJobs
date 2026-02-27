package controller

import (
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AccountController struct {
	userRepo interfaces.UserRepositoryInterface
}

func NewAccountController(userRepo interfaces.UserRepositoryInterface) *AccountController {
	return &AccountController{userRepo: userRepo}
}

type deleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

func (ac *AccountController) DeleteAccount(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido"})
		return
	}

	var body deleteAccountRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Senha é obrigatória para excluir conta"})
		return
	}

	// Fetch full user with password hash
	fullUser, err := ac.userRepo.GetUserByEmail(user.Email)
	if err != nil || fullUser.Id == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar usuário"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(fullUser.Password), []byte(body.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	if err := ac.userRepo.SoftDeleteUser(user.Id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir conta"})
		return
	}

	// Clear auth cookie
	ctx.SetCookie("Authorization", "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "Conta excluída com sucesso"})
}
