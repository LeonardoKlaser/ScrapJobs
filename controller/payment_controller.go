package controller

import (
	"net/http"
	"strconv"
	"web-scrapper/gateway" // Importe o DTO do webhook
	"web-scrapper/interfaces"
	"web-scrapper/logging" // Seu logger
	"web-scrapper/model"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	paymentUsecase *usecase.PaymentUsecase
	planUsecase *usecase.PlanUsecase
	emailService interfaces.EmailService
}

func NewPaymentcontroller(paymentUsecase *usecase.PaymentUsecase, planUsecase *usecase.PlanUsecase, es interfaces.EmailService) *PaymentController{
	return &PaymentController{
		paymentUsecase: paymentUsecase,
		planUsecase: planUsecase,
		emailService: es,
	}
}

type CreatePaymentRequest struct {
	Methods   []string `json:"methods" binding:"required"`  
	Frequency string   `json:"frequency" binding:"required"` 
}

func (p *PaymentController) CreatePayment(ctx *gin.Context) {
	planId, err := strconv.Atoi(ctx.Param("planId"))
	if err != nil{
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "Falha recuperar identificador do plano"})
	}
    
	plan, err := p.planUsecase.GetPlanByID(planId)
	if err != nil{
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "Falha ao buscar plano"})
	}

	var reqBody gateway.InitiatePaymentRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		logging.Logger.Warn().Err(err).Msg("Payload de iniciação de pagamento inválido")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido: " + err.Error()})
		return
	}

    paymentURL, err := p.paymentUsecase.CreatePayment(ctx.Request.Context(), *plan, reqBody)
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
		logging.Logger.Warn().Err(err).Msg("Webhook AbacatePay: Payload inválido")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	// --- !!! IMPLEMENTAR VERIFICAÇÃO DE ASSINATURA AQUI !!! ---
	// signature := ctx.GetHeader("X-Abacatepay-Signature") // Ou o header correto
	// webhookSecret := os.Getenv("ABACATEPAY_WEBHOOK_SECRET")
	// if !gateway.VerifyWebhookSignature(signature, ctx.Request.Body, webhookSecret) {
	//     logging.Logger.Error().Msg("Webhook AbacatePay: Assinatura inválida")
	//     ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Assinatura inválida"})
	//     return
	// }

	logging.Logger.Info().Str("event", payload.Event).Str("external_ref", payload.Data.Object.ExternalReference).Msg("Webhook AbacatePay recebido")

	// Processar apenas eventos de pagamento confirmado
	if payload.Event == "billing.paid" || payload.Data.Object.Status == "PAID" { // Verifique o evento/status correto
		
		pendingRegistrationID := payload.Data.Object.ExternalReference
		if pendingRegistrationID == "" {
			logging.Logger.Error().Msg("Webhook AbacatePay: ExternalReference ausente no payload 'billing.paid'")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Referência externa ausente"})
			return
		}

		// Chamar usecase para finalizar o registro
		newUser, err := p.paymentUsecase.CompleteRegistration(ctx.Request.Context(), pendingRegistrationID)
		if err != nil {
			logging.Logger.Error().Err(err).Str("pending_id", pendingRegistrationID).Msg("Webhook AbacatePay: Falha ao completar registro")
			// Retornar 200 OK mesmo assim para evitar retentativas da AbacatePay por este erro?
			// Ou retornar 500 para AbacatePay tentar de novo? Depende da sua estratégia.
			// Por segurança, retornar 200 e logar bem o erro é mais seguro para não criar usuários duplicados.
			ctx.JSON(http.StatusOK, gin.H{"status": "received_but_failed_processing"})
			return
		}

		logging.Logger.Info().Int("user_id", newUser.Id).Str("email", newUser.Email).Msg("Webhook AbacatePay: Usuário criado com sucesso")

		// Enviar e-mail de boas-vindas
		// TODO: Criar o método SendWelcomeEmail
		// dashboardLink := "http://localhost:5173/dashboard" // Ou a URL correta
		// err = p.emailService.SendWelcomeEmail(newUser.Email, newUser.Name, dashboardLink)
		// if err != nil {
		//     // Logar erro de email, mas não falhar o webhook por isso
		//     logging.Logger.Error().Err(err).Int("user_id", newUser.Id).Msg("Falha ao enviar e-mail de boas-vindas")
		// }

	} else {
		logging.Logger.Info().Str("event", payload.Event).Msg("Webhook AbacatePay: Evento ignorado")
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "received"})
}