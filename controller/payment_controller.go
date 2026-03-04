package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"web-scrapper/gateway"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/tasks"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type PaymentController struct {
	paymentUsecase *usecase.PaymentUsecase
	emailService   interfaces.EmailService
	asynqClient    *asynq.Client
}

func NewPaymentController(
	pu *usecase.PaymentUsecase,
	es interfaces.EmailService,
	ac *asynq.Client,
) *PaymentController {
	return &PaymentController{
		paymentUsecase: pu,
		emailService:   es,
		asynqClient:    ac,
	}
}

// CreatePayment godoc
// @Summary Criar pagamento
// @Description Inicia processo de pagamento para um plano
// @Tags Payment
// @Accept json
// @Produce json
// @Param planId path int true "ID do plano"
// @Param body body gateway.InitiatePaymentRequest true "Dados de pagamento"
// @Success 200 {object} model.CreatePaymentResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Failure 502 {object} model.ErrorResponse
// @Router /api/payments/create/{planId} [post]
func (p *PaymentController) CreatePayment(ctx *gin.Context) {
	log := logging.Logger
	planIDStr := ctx.Param("planId")
	planID, err := strconv.Atoi(planIDStr)
	if err != nil {
		log.Warn().Err(err).Str("plan_id_str", planIDStr).Msg("ID do plano inválido na URL")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID do plano inválido"})
		return
	}

	var reqBody gateway.InitiatePaymentRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		log.Warn().Err(err).Msg("Payload de iniciação de pagamento inválido")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido: " + err.Error()})
		return
	}

	// Limpa CPF e celular (remove caracteres não numéricos)
	reqBody.Tax = cleanString(reqBody.Tax)
	reqBody.Cellphone = cleanString(reqBody.Cellphone)

	log.Info().Str("email", reqBody.Email).Int("plan_id", planID).Msg("Iniciando processo de pagamento")
	paymentURL, err := p.paymentUsecase.CreatePayment(ctx.Request.Context(), planID, reqBody)
	if err != nil {
		log.Error().Err(err).Str("email", reqBody.Email).Int("plan_id", planID).Msg("Erro ao iniciar pagamento no usecase")
		friendlyMsg, statusCode := parsePaymentError(err.Error())
		ctx.JSON(statusCode, gin.H{"error": friendlyMsg})
		return
	}

	log.Info().Str("email", reqBody.Email).Int("plan_id", planID).Msg("URL de pagamento gerada com sucesso")
	ctx.JSON(http.StatusOK, gin.H{"url": paymentURL})
}

// HandleWebhook godoc
// @Summary Webhook de pagamento
// @Description Recebe notificacao de pagamento do AbacatePay
// @Tags Payment
// @Accept json
// @Produce json
// @Param body body gateway.WebhookPayload true "Payload do webhook"
// @Success 200 {object} model.StatusResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/webhooks/abacatepay [post]
func (p *PaymentController) HandleWebhook(ctx *gin.Context) {
	log := logging.Logger

	// Lê o corpo bruto para verificação HMAC e parse
	rawBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Webhook AbacatePay: Erro ao ler corpo da requisição")
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error reading body"})
		return
	}
	// Restaura o body para uso posterior se necessário
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var payload gateway.WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		log.Warn().Err(err).Bytes("raw_body", rawBody).Msg("Webhook AbacatePay: Payload JSON inválido")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	log.Info().Str("event", payload.Event).Str("webhook_id", payload.ID).Bool("dev_mode", payload.DevMode).Msg("Webhook AbacatePay recebido")

	// Processa apenas o evento billing.paid
	if payload.Event != "billing.paid" {
		log.Info().Str("event", payload.Event).Msg("Webhook AbacatePay: Evento ignorado (não é billing.paid)")
		ctx.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	billing := payload.Data.Billing
	if billing == nil {
		log.Error().Str("webhook_id", payload.ID).Msg("Webhook AbacatePay: Campo 'data.billing' ausente no evento billing.paid")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payload incompleto: billing ausente"})
		return
	}

	if billing.Status != "PAID" {
		log.Info().Str("status", billing.Status).Msg("Webhook AbacatePay: Billing não está PAID, ignorando")
		ctx.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	// Extrai email do customer.metadata (chave Redis agora é baseada no email)
	customerEmail := ""
	if billing.Customer != nil {
		customerEmail = billing.Customer.Metadata["email"]
	}
	if customerEmail == "" {
		log.Error().Str("webhook_id", payload.ID).Str("billing_id", billing.ID).Msg("Webhook AbacatePay: email ausente no customer.metadata")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email do cliente ausente no webhook"})
		return
	}

	log.Info().
		Str("webhook_id", payload.ID).
		Str("billing_id", billing.ID).
		Str("customer_email", customerEmail).
		Msg("Enfileirando task para completar registro do usuário")

	taskPayload, err := json.Marshal(tasks.CompleteRegistrationPayload{
		PendingRegistrationID: customerEmail,
		CustomerEmail:         customerEmail,
	})
	if err != nil {
		log.Error().Err(err).Str("customer_email", customerEmail).Msg("Falha ao serializar payload da task CompleteRegistration")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao preparar processamento"})
		return
	}

	task := asynq.NewTask(tasks.TypeCompleteRegistration, taskPayload, asynq.MaxRetry(5))
	info, err := p.asynqClient.Enqueue(task, asynq.Queue("critical"))
	if err != nil {
		log.Error().Err(err).Str("customer_email", customerEmail).Msg("Falha ao enfileirar task CompleteRegistration")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao agendar processamento"})
		return
	}

	log.Info().
		Str("task_id", info.ID).
		Str("customer_email", customerEmail).
		Msg("Task CompleteRegistration enfileirada com sucesso")

	ctx.JSON(http.StatusOK, gin.H{"status": "received"})
}

func parsePaymentError(errMsg string) (string, int) {
	lower := strings.ToLower(errMsg)
	switch {
	case strings.Contains(lower, "taxid") || strings.Contains(lower, "cpf"):
		return "CPF/CNPJ inválido. Verifique o número informado.", http.StatusBadRequest
	case strings.Contains(lower, "email"):
		return "E-mail inválido. Verifique o endereço informado.", http.StatusBadRequest
	case strings.Contains(lower, "cellphone") || strings.Contains(lower, "phone"):
		return "Telefone inválido. Verifique o número informado.", http.StatusBadRequest
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline"):
		return "Serviço de pagamento temporariamente indisponível. Tente novamente.", http.StatusBadGateway
	default:
		return "Erro ao processar pagamento. Tente novamente.", http.StatusInternalServerError
	}
}

// cleanString remove todos os caracteres não numéricos (para CPF, celular, etc.)
func cleanString(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
