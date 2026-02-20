package controller

import (
	"net/http"
	"os"
	"time"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is used to prevent timing-based user enumeration
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

type UserController struct {
	usecase *usecase.UserUsecase
}

func NewUserController(usercase *usecase.UserUsecase) UserController {
	return UserController{
		usecase: usercase,
	}
}

func (usr *UserController) SignUp(ctx *gin.Context) {
	var body struct {
		Name      string  `json:"user_name"`
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		Tax       *string `json:"tax,omitempty"`
		Cellphone *string `json:"cellphone,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := model.User{
		Name:      body.Name,
		Email:     body.Email,
		Password:  body.Password,
		Tax:       body.Tax,
		Cellphone: body.Cellphone,
	}

	_, err := usr.usecase.CreateUser(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, nil)
}

func (usr *UserController) SignIn(ctx *gin.Context) {
	var body struct {
		Email    string
		Password string
	}

	if ctx.Bind(&body) != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	if body.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User email is required in the path"})
		return
	}

	res, err := usr.usecase.GetUserByEmail(body.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Timing-safe: always run bcrypt even if user not found
	hashToCompare := []byte(res.Password)
	if res.Id == 0 {
		hashToCompare = dummyHash
	}

	err = bcrypt.CompareHashAndPassword(hashToCompare, []byte(body.Password))
	if err != nil || res.Id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid E-mail or Password",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": res.Id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWTTOKEN")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	isSecure := os.Getenv("GIN_MODE") == "release"
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", tokenString, 3600*24, "", "", isSecure, true)

	ctx.JSON(http.StatusOK, gin.H{})
}

func (usr *UserController) Logout(ctx *gin.Context) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", "", -1, "", "", isSecure, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (usr *UserController) UpdateProfile(ctx *gin.Context) {
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

	var body struct {
		Name      string  `json:"user_name" binding:"required"`
		Cellphone *string `json:"cellphone"`
		Tax       *string `json:"tax"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := usr.usecase.UpdateUserProfile(user.Id, body.Name, body.Cellphone, body.Tax)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Perfil atualizado com sucesso"})
}

func (usr *UserController) ChangePassword(ctx *gin.Context) {
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

	var body struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.NewPassword) < 8 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "A nova senha deve ter pelo menos 8 caracteres"})
		return
	}

	err := usr.usecase.ChangePassword(user.Id, user.Password, body.OldPassword, body.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Senha alterada com sucesso"})
}
