package controller

import (
	"fmt"
	"net/http"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type CurriculumController struct {
	curriculumUsecase usecase.CurriculumUsecase
}

func NewCurriculumController (usecase usecase.CurriculumUsecase) CurriculumController{
	return CurriculumController{
		curriculumUsecase: usecase,
	}
}

func (c *CurriculumController) CreateCurriculum(ctx *gin.Context) {
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

	var curriculum model.Curriculum
	if err := ctx.ShouldBindJSON(&curriculum); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error to deserialize new job json body: %w", err)})
	}

	curriculum.UserID = user.Id

	res, err := c.curriculumUsecase.CreateCurriculum(curriculum)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}


func (c *CurriculumController) GetCurriculumByUserId(ctx *gin.Context) {
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

	res, err := c.curriculumUsecase.GetCurriculumByUserId(user.Id)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}
