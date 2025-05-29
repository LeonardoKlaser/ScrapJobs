package controller

import (
	"fmt"
	"web-scrapper/usecase"
	"web-scrapper/model"
	"net/http"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	usecase usecase.UserUsecase
}

func NewUserController(usercase usecase.UserUsecase) UserController{
	return UserController{
		usecase: usercase,
	}
}

func (usr *UserController) CreateUser(ctx *gin.Context) {
	var user model.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error" : fmt.Errorf("error to get new user json: %w", err )})
		return
	}

	res, err := usr.usecase.CreateUser(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}


func (usr *UserController) GetUserByEmail(ctx *gin.Context) {
	userEmail := ctx.Param("email")

	if userEmail == ""{
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User email is required in the path"})
		return
	}

	res, err := usr.usecase.GetUserByEmail(userEmail)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}