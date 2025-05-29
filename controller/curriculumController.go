package controller

import (
	"fmt"
	"net/http"
	"web-scrapper/model"
	"web-scrapper/usecase"
    "strconv"
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
	var curriculum model.Curriculum
	if err := ctx.ShouldBindJSON(&curriculum); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error to deserialize new job json body: %w", err)})
	}

	res, err := c.curriculumUsecase.CreateCurriculum(curriculum)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}


func (c *CurriculumController) GetCurriculumByUserId(ctx *gin.Context) {
	userId := ctx.Param("id")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error to deserialize new job json body: %w", err)})
	}

	res, err := c.curriculumUsecase.GetCurriculumByUserId(userIdInt)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}
