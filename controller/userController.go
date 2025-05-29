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
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("error to get new user json: %w", err ))
		return
	}

	res, err := usr.usecase.CreateUser(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}


func (usr *UserController) GetUserByEmail(ctx *gin.Context) {
	var userEmail string

	if err := ctx.ShouldBindJSON(&userEmail); err != nil {
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("error to get email json: %w", err ))
		return
	}

	res, err := usr.usecase.GetUserByEmail(userEmail)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, res)
}