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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar processo de pagamento"})
		return
	}

	log.Info().Str("email", reqBody.Email).Int("plan_id", planID).Msg("URL de pagamento gerada com sucesso")
	ctx.JSON(http.StatusOK, gin.H{"url": paymentURL})
}

// HandleWebhook processa os eventos de webhook da AbacatePay.
// Suporta o evento "billing.paid" para completar o registro do usuário.
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

	// O ExternalId do billing é o nosso pendingRegistrationID
	pendingRegistrationID := billing.ExternalId
	if pendingRegistrationID == "" {
		log.Error().Str("webhook_id", payload.ID).Str("billing_id", billing.ID).Msg("Webhook AbacatePay: ExternalId ausente no billing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ExternalId ausente no billing"})
		return
	}

	log.Info().
		Str("webhook_id", payload.ID).
		Str("billing_id", billing.ID).
		Str("pending_reg_id", pendingRegistrationID).
		Msg("Enfileirando task para completar registro do usuário")

	taskPayload, err := json.Marshal(tasks.CompleteRegistrationPayload{
		PendingRegistrationID: pendingRegistrationID,
		CustomerEmail:         billing.Customer.Metadata["email"],
	})
	if err != nil {
		log.Error().Err(err).Str("pending_reg_id", pendingRegistrationID).Msg("Falha ao serializar payload da task CompleteRegistration")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao preparar processamento"})
		return
	}

	task := asynq.NewTask(tasks.TypeCompleteRegistration, taskPayload, asynq.MaxRetry(5))
	info, err := p.asynqClient.Enqueue(task, asynq.Queue("critical"))
	if err != nil {
		log.Error().Err(err).Str("pending_reg_id", pendingRegistrationID).Msg("Falha ao enfileirar task CompleteRegistration")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao agendar processamento"})
		return
	}

	log.Info().
		Str("task_id", info.ID).
		Str("pending_reg_id", pendingRegistrationID).
		Msg("Task CompleteRegistration enfileirada com sucesso")

	ctx.JSON(http.StatusOK, gin.H{"status": "received"})
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
