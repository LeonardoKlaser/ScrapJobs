package controller

import (
	"errors"
	"fmt"
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CheckAuthController struct {
	userRepo interfaces.UserRepositoryInterface
}

func NewCheckAuthController(userRepo interfaces.UserRepositoryInterface) CheckAuthController {
	return CheckAuthController{
		userRepo: userRepo,
	}
}

// CheckAuthUser godoc
// @Summary Verificar autenticacao
// @Description Retorna dados do usuario autenticado (id/email/is_admin do JWT + dados dinâmicos do banco)
// @Tags Auth
// @Produce json
// @Success 200 {object} object
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/me [get]
func (controller *CheckAuthController) CheckAuthUser(ctx *gin.Context) {
	claimsInterface, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	claims, ok := claimsInterface.(jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Claims inválidos"})
		return
	}

	// Extract user ID from JWT claims
	var userID int
	switch v := claims["sub"].(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	case string:
		fmt.Sscanf(v, "%d", &userID)
	default:
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// Only id, email, and is_admin come from JWT (immutable data).
	// user_name, cellphone, tax, etc. come from the DB to stay fresh after profile updates.
	email, _ := claims["email"].(string)
	isAdmin, _ := claims["is_admin"].(bool)

	// Fetch dynamic + editable data from DB (single lightweight query)
	meData, err := controller.userRepo.GetUserMeData(userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Conta desativada"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados do usuário"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":                     userID,
		"user_name":              meData.UserName,
		"email":                  email,
		"tax":                    meData.Tax,
		"cellphone":              meData.Cellphone,
		"is_admin":               isAdmin,
		"plan":                   meData.Plan,
		"expires_at":             meData.ExpiresAt,
		"monitored_sites_count":  meData.MonitoredSitesCount,
		"monthly_analysis_count": meData.MonthlyAnalysisCount,
		"weekdays_only":          meData.WeekdaysOnly,
	})
}
