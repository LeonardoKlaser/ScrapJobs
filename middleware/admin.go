package middleware

import (
	"net/http"
	"os"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userInterface, exists := ctx.Get("user")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			return
		}

		user, ok := userInterface.(model.User)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido no contexto"})
			return
		}

		adminEmail := os.Getenv("ADMIN_EMAIL")
		if !user.IsAdmin && user.Email != adminEmail {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
			return
		}

		ctx.Next()
	}
}
