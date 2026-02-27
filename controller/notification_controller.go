package controller

import (
	"net/http"
	"strconv"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	notificationUsecase *usecase.NotificationsUsecase
}

func NewNotificationController(uc *usecase.NotificationsUsecase) *NotificationController {
	return &NotificationController{
		notificationUsecase: uc,
	}
}

// GetNotificationsByUser godoc
// @Summary Listar notificacoes
// @Description Retorna historico de notificacoes do usuario
// @Tags Notifications
// @Produce json
// @Param limit query int false "Limite de resultados (max 200)" default(50)
// @Success 200 {array} model.NotificationWithJob
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/notifications [get]
func (nc *NotificationController) GetNotificationsByUser(ctx *gin.Context) {
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

	limit := 50
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	notifications, err := nc.notificationUsecase.GetNotificationsByUser(user.Id, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notifications == nil {
		notifications = []model.NotificationWithJob{}
	}

	ctx.JSON(http.StatusOK, notifications)
}
