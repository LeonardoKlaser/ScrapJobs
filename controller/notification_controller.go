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

// GetNotificationsByUser retorna o histórico de notificações do usuário autenticado.
// Query param opcional: limit (default: 50, max: 200)
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
