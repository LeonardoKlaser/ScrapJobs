package middleware

import (
	"fmt"
	"net/http"
	"os"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	user *usecase.UserUsecase
}

func NewMiddleware(user *usecase.UserUsecase) *Middleware {
	return &Middleware{
		user: user,
	}
}

func (m *Middleware) RequireAuth(ctx *gin.Context) {
	tokenString, err := ctx.Cookie("Authorization")

	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signin method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWTTOKEN")), nil
	})

	if err != nil || token == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var userID int
	switch v := claims["sub"].(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	case string:
		fmt.Sscanf(v, "%d", &userID)
	default:
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, err := m.user.GetUserById(userID)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.Id == 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set("user", user)

	ctx.Next()
}
