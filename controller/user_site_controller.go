package controller

import (
	"fmt"
	"net/http"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)


type UserSiteController struct{
	usecase *usecase.UserSiteUsecase
}

func NewUserSiteController(usecase *usecase.UserSiteUsecase) *UserSiteController{
	return &UserSiteController{
		usecase: usecase,
	}
}

// InsertUserSite godoc
// @Summary Inscrever-se em site
// @Description Adiciona inscricao do usuario em um site de carreiras
// @Tags UserSite
// @Accept json
// @Produce json
// @Param body body model.UserSiteRequest true "ID do site e palavras-chave"
// @Success 201 {object} object
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /userSite [post]
func (usecase *UserSiteController) InsertUserSite(ctx *gin.Context) {
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

	var body model.UserSiteRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos para inscrição no site"})
		return
	}

	err := usecase.usecase.InsertUserSite(user.Id, body.SiteId, body.TargetWords)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao inscrever usuário no site"})
		return
	}


	ctx.JSON(http.StatusCreated, gin.H{})
}

// DeleteUserSite godoc
// @Summary Cancelar inscricao em site
// @Description Remove inscricao do usuario em um site de carreiras
// @Tags UserSite
// @Produce json
// @Param siteId path string true "ID do site"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /userSite/{siteId} [delete]
func (usc *UserSiteController) DeleteUserSite(ctx *gin.Context) {
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

	siteId := ctx.Param("siteId")
	if siteId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "siteId é obrigatório"})
		return
	}

	err := usc.usecase.DeleteUserSite(user.Id, siteId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Inscrição removida com sucesso"})
}

// UpdateUserSiteFilters godoc
// @Summary Atualizar filtros do site
// @Description Atualiza as palavras-chave de filtragem para um site inscrito
// @Tags UserSite
// @Accept json
// @Produce json
// @Param siteId path string true "ID do site"
// @Param body body model.UpdateUserSiteFiltersRequest true "Palavras-chave"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /userSite/{siteId} [patch]
func (usc *UserSiteController) UpdateUserSiteFilters(ctx *gin.Context) {
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

	siteIdStr := ctx.Param("siteId")
	if siteIdStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "siteId é obrigatório"})
		return
	}

	var siteId int
	if _, err := fmt.Sscanf(siteIdStr, "%d", &siteId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "siteId deve ser um número inteiro"})
		return
	}

	var body struct {
		TargetWords []string `json:"target_words"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido: " + err.Error()})
		return
	}

	if err := usc.usecase.UpdateUserSiteFilters(user.Id, siteId, body.TargetWords); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Filtros atualizados com sucesso"})
}
