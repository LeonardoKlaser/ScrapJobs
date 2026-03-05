package controller

import (
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type EmailConfigController struct {
	repo         interfaces.EmailConfigRepository
	orchestrator *usecase.EmailOrchestrator
}

func NewEmailConfigController(repo interfaces.EmailConfigRepository, orchestrator *usecase.EmailOrchestrator) *EmailConfigController {
	return &EmailConfigController{repo: repo, orchestrator: orchestrator}
}

func (c *EmailConfigController) GetEmailConfig(ctx *gin.Context) {
	configs, err := c.repo.GetAll()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao buscar configuração de email")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar configuração"})
		return
	}
	ctx.JSON(http.StatusOK, configs)
}

type UpdateEmailConfigRequest struct {
	Providers []struct {
		ProviderName string `json:"provider_name" binding:"required"`
		IsActive     bool   `json:"is_active"`
		Priority     int    `json:"priority" binding:"required"`
	} `json:"providers" binding:"required"`
}

func (c *EmailConfigController) UpdateEmailConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var req UpdateEmailConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Validate at least one provider is active
	hasActive := false
	for _, p := range req.Providers {
		if p.IsActive {
			hasActive = true
			break
		}
	}
	if !hasActive {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Pelo menos um provedor deve estar ativo"})
		return
	}

	configs := make([]model.EmailProviderConfig, len(req.Providers))
	for i, p := range req.Providers {
		configs[i] = model.EmailProviderConfig{
			ProviderName: p.ProviderName,
			IsActive:     p.IsActive,
			Priority:     p.Priority,
		}
	}

	uid := userID.(int)
	if err := c.repo.Update(configs, uid); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao atualizar configuração de email")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar configuração"})
		return
	}

	c.orchestrator.InvalidateCache()

	logging.Logger.Info().Int("updated_by", uid).Msg("Configuração de email atualizada")
	ctx.JSON(http.StatusOK, gin.H{"message": "Configuração atualizada com sucesso"})
}
