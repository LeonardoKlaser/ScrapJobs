package middleware

import (
	"net/http"
	"time"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

func RequireActiveSubscription() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userInterface, exists := ctx.Get("user")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			return
		}
		user, ok := userInterface.(model.User)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido"})
			return
		}
		if user.ExpiresAt != nil && user.ExpiresAt.Before(time.Now()) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "subscription_expired"})
			return
		}
		ctx.Next()
	}
}
