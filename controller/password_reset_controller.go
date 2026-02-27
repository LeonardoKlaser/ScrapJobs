package controller

import (
	"net/http"
	"os"
	"time"
	"web-scrapper/interfaces"
	"web-scrapper/logging"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type PasswordResetController struct {
	resetRepo interfaces.PasswordResetRepositoryInterface
	userRepo  interfaces.UserRepositoryInterface
	emailSvc  interfaces.EmailService
}

func NewPasswordResetController(
	resetRepo interfaces.PasswordResetRepositoryInterface,
	userRepo interfaces.UserRepositoryInterface,
	emailSvc interfaces.EmailService,
) *PasswordResetController {
	return &PasswordResetController{
		resetRepo: resetRepo,
		userRepo:  userRepo,
		emailSvc:  emailSvc,
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (c *PasswordResetController) ForgotPassword(ctx *gin.Context) {
	var body forgotPasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "E-mail invalido"})
		return
	}

	// Always return success to prevent email enumeration
	user, err := c.userRepo.GetUserByEmail(body.Email)
	if err != nil || user.Id == 0 {
		ctx.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviaremos instrucoes de recuperacao."})
		return
	}

	token, err := c.resetRepo.CreateToken(user.Id, 1*time.Hour)
	if err != nil {
		logging.Logger.Error().Err(err).Str("email", body.Email).Msg("Erro ao criar token de recuperacao")
		ctx.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviaremos instrucoes de recuperacao."})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := frontendURL + "/reset-password?token=" + token

	if err := c.emailSvc.SendPasswordResetEmail(ctx.Request.Context(), user.Email, user.Name, resetLink); err != nil {
		logging.Logger.Error().Err(err).Str("email", body.Email).Msg("Erro ao enviar email de recuperacao")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviaremos instrucoes de recuperacao."})
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (c *PasswordResetController) ResetPassword(ctx *gin.Context) {
	var body resetPasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token e nova senha (minimo 8 caracteres) sao obrigatorios"})
		return
	}

	tokenRecord, err := c.resetRepo.FindValidToken(body.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}
	if tokenRecord == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token invalido ou expirado"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar nova senha"})
		return
	}

	if err := c.userRepo.UpdateUserPassword(tokenRecord.UserID, string(hashed)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar senha"})
		return
	}

	if err := c.resetRepo.MarkUsed(tokenRecord.ID); err != nil {
		logging.Logger.Error().Err(err).Int("token_id", tokenRecord.ID).Msg("Erro ao marcar token como usado")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Senha atualizada com sucesso"})
}
