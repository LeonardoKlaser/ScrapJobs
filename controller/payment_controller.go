package controller

import (
	"net/http"
	"strconv"
	"web-scrapper/gateway" // Importe o DTO do webhook
	"web-scrapper/logging" // Seu logger
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	paymentUsecase *usecase.PaymentUsecase
	planUsecase *usecase.PlanUsecase
}

func NewPaymentcontroller(paymentUsecase *usecase.PaymentUsecase, planUsecase *usecase.PlanUsecase) *PaymentController{
	return &PaymentController{
		paymentUsecase: paymentUsecase,
		planUsecase: planUsecase,
	}
}

type CreatePaymentRequest struct {
	Methods   []string `json:"methods" binding:"required"`   // Ex: ["PIX", "CREDIT_CARD"]
	Frequency string   `json:"frequency" binding:"required"` // Ex: "ONE_TIME" ou "MULTIPLE_PAYMENTS"
}

func (p *PaymentController) CreatePayment(ctx *gin.Context) {
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

	siteId, err := strconv.Atoi(ctx.Param("planId"))
	if err != nil{
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "Falha recuperar identificador do plano"})
	}
    
	plan, err := p.planUsecase.GetPlanByID(siteId)
	if err != nil{
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "Falha ao buscar plano"})
	}

	var reqBody CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		logging.Logger.Warn().Err(err).Msg("Payload de criação de pagamento inválido")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido: " + err.Error()})
		return
	}

    paymentURL, err := p.paymentUsecase.CreatePayment(ctx.Request.Context(), *plan, user, reqBody.Methods, reqBody.Frequency)
    if err != nil {
		logging.Logger.Error().Err(err).Msg("Erro ao criar cobrança")
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar cobrança"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"url": paymentURL})
}



func (p *PaymentController) HandleWebhook(ctx *gin.Context) {
	var payload gateway.WebhookPayload

	
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logging.Logger.Warn().Err(err).Msg("Webhook da AbacatePay com payload inválido")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	// IMPORTANTE: Verificação de Assinatura do Webhook
	// A AbacatePay deve enviar um cabeçalho (ex: `X-Signature`)
	// Você DEVE verificar essa assinatura usando seu "Webhook Secret"
	// para garantir que a requisição veio da AbacatePay.
	// Por agora, vamos pular, mas anote para implementar:
	// if !gateway.VerifySignature(ctx.GetHeader("X-Signature"), ctx.Request.Body) {
	//     ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Assinatura inválida"})
	//     return
	// }
	
	logging.Logger.Info().Str("event", payload.Event).Msg("Webhook da AbacatePay recebido")

	// 2. Processar o evento
	switch payload.Event {
	case "billing.paid":
		// É AQUI QUE A MÁGICA ACONTECE!
		// Pegue os dados do payload.Data.Object...
		// customerEmail := payload.Data.Object.CustomerEmail
		
		// Encontre o usuário no seu banco de dados
		// user, err := p.userUsecase.FindByEmail(customerEmail)
		
		// Atualize o status da assinatura dele
		// err := p.userUsecase.ActivatePlan(user.ID)
		
		logging.Logger.Info().Msg("Plano ativado com sucesso via webhook.")

	case "billing.failed":
		// Lógica para pagamento falho
		logging.Logger.Warn().Msg("Pagamento falhou via webhook.")

	// ... outros eventos (billing.canceled, etc.)
	}

	// 3. Responder 200 OK para a AbacatePay saber que recebemos
	ctx.JSON(http.StatusOK, gin.H{"status": "received"})
}