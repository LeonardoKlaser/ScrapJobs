package controller

import (
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

type CheckAuthController struct {
	userSiteRepo interfaces.UserSiteRepositoryInterface
}

func NewCheckAuthController(userSiteRepo interfaces.UserSiteRepositoryInterface) CheckAuthController {
	return CheckAuthController{
		userSiteRepo: userSiteRepo,
	}
}

// CheckAuthUser godoc
// @Summary Verificar autenticacao
// @Description Retorna dados do usuario autenticado
// @Tags Auth
// @Produce json
// @Success 200 {object} model.User
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/me [get]
func (controller *CheckAuthController) CheckAuthUser(ctx *gin.Context) {
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

	siteCount, err := controller.userSiteRepo.GetUserSiteCount(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados do usuário"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_name":             user.Name,
		"email":                 user.Email,
		"tax":                   user.Tax,
		"cellphone":             user.Cellphone,
		"is_admin":              user.IsAdmin,
		"plan":                  user.Plan,
		"expires_at":            user.ExpiresAt,
		"monitored_sites_count": siteCount,
	})
}
