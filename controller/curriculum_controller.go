package controller

import (
	"net/http"
	"strconv"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type CurriculumController struct {
	curriculumUsecase *usecase.CurriculumUsecase
}

func NewCurriculumController (usecase *usecase.CurriculumUsecase) *CurriculumController{
	return &CurriculumController{
		curriculumUsecase: usecase,
	}
}

// CreateCurriculum godoc
// @Summary Criar curriculo
// @Description Cria um novo curriculo para o usuario autenticado
// @Tags Curriculum
// @Accept json
// @Produce json
// @Param body body model.Curriculum true "Dados do curriculo"
// @Success 201 {object} model.Curriculum
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum [post]
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	curriculum.UserID = user.Id

	res, err := c.curriculumUsecase.CreateCurriculum(curriculum)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}

// GetCurriculumByUserId godoc
// @Summary Listar curriculos
// @Description Retorna todos os curriculos do usuario autenticado
// @Tags Curriculum
// @Produce json
// @Success 200 {array} model.Curriculum
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum [get]
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

	ctx.JSON(http.StatusOK, res)
}

// UpdateCurriculum godoc
// @Summary Atualizar curriculo
// @Description Atualiza um curriculo existente
// @Tags Curriculum
// @Accept json
// @Produce json
// @Param id path int true "ID do curriculo"
// @Param body body model.Curriculum true "Dados atualizados"
// @Success 200 {object} model.Curriculum
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum/{id} [put]
func (c *CurriculumController) UpdateCurriculum(ctx *gin.Context) {
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	curriculum.UserID = user.Id

	res, err := c.curriculumUsecase.UpdateCurriculum(curriculum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// SetActiveCurriculum godoc
// @Summary Ativar curriculo
// @Description Define um curriculo como ativo para o usuario
// @Tags Curriculum
// @Produce json
// @Param id path int true "ID do curriculo"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum/{id}/active [patch]
func (c *CurriculumController) SetActiveCurriculum(ctx *gin.Context) {
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

	curriculumID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de currículo inválido"})
		return
	}

	err = c.curriculumUsecase.SetActiveCurriculum(user.Id, curriculumID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Currículo ativado com sucesso"})
}
