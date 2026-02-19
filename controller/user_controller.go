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

	err = bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(body.Password))

	if err != nil {
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

	ctx.SetSameSite(http.SameSiteDefaultMode)
	ctx.SetCookie("Authorization", tokenString, 3600*24, "", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{})
}

func (usr *UserController) Logout(ctx *gin.Context) {
    ctx.SetCookie("Authorization", "", -1, "", "", true, true)
    ctx.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
