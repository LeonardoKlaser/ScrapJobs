package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"web-scrapper/gateway"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/tasks"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// Ensure gateway.AbacatePayGateway satisfies the interface at compile time
var _ interfaces.PaymentGatewayInterface = (*gateway.AbacatePayGateway)(nil)

type PaymentUsecase struct {
	paymentGateway interfaces.PaymentGatewayInterface
	redisClient    redis.UniversalClient
	userUsecase    *UserUsecase
	planRepository interfaces.PlanRepositoryInterface
}

func NewPaymentUsecase(gw interfaces.PaymentGatewayInterface, redisClient redis.UniversalClient, userUsecase *UserUsecase, planRp interfaces.PlanRepositoryInterface) *PaymentUsecase {
	return &PaymentUsecase{paymentGateway: gw, redisClient: redisClient, userUsecase: userUsecase, planRepository: planRp}
}

// PixQRCodeResult contém os dados retornados ao frontend após criação do QR Code.
type PixQRCodeResult struct {
	PixID        string `json:"pix_id"`
	BrCode       string `json:"br_code"`
	BrCodeBase64 string `json:"br_code_base64"`
	ExpiresAt    string `json:"expires_at"`
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, planID int, userData gateway.InitiatePaymentRequest) (*PixQRCodeResult, error) {
	log := logging.Logger.With().Str("usecase", "PaymentUsecase").Str("method", "CreatePayment").Logger()

	plan, err := uc.planRepository.GetPlanByID(planID)
	if err != nil || plan == nil {
		if err == nil && plan == nil {
			err = errors.New("plano não encontrado")
		}
		log.Error().Err(err).Int("plan_id", planID).Msg("Erro ao buscar plano")
		return nil, fmt.Errorf("plano com ID %d não encontrado: %w", planID, err)
	}
	log.Info().Int("plan_id", planID).Str("plan_name", plan.Name).Msg("Plano encontrado")

	// Limpa CPF e celular antes de armazenar
	userData.Tax = cleanNumericString(userData.Tax)
	userData.Cellphone = cleanNumericString(userData.Cellphone)

	// Hash password before storing in Redis to avoid plaintext exposure
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao gerar hash da senha")
		return nil, fmt.Errorf("erro ao processar senha: %w", err)
	}

	// Calcula preço final com base no período de cobrança
	var finalPrice float64
	var productName string
	if userData.BillingPeriod == "quarterly" {
		finalPrice = plan.Price * 3 * 0.85 // 15% discount
		productName = plan.Name + " (Trimestral)"
	} else {
		finalPrice = plan.Price
		productName = plan.Name + " (Mensal)"
	}
	priceInCents := int(math.Round(finalPrice * 100))

	// Trunca description para 37 caracteres (limite da AbacatePay)
	description := "ScrapJobs - " + productName
	if len(description) > 37 {
		description = description[:37]
	}

	customer := &gateway.PixCustomer{
		Name:      userData.Name,
		Email:     userData.Email,
		Cellphone: userData.Cellphone,
		TaxId:     userData.Tax,
	}

	log.Info().Str("email", userData.Email).Int("plan_id", planID).Int("amount_cents", priceInCents).Msg("Criando QR Code PIX na AbacatePay")
	pixData, err := uc.paymentGateway.CreatePixQRCode(ctx, priceInCents, 900, description, customer)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao criar QR Code PIX na AbacatePay")
		return nil, fmt.Errorf("erro ao criar QR Code PIX: %w", err)
	}
	log.Info().Str("pix_id", pixData.ID).Msg("QR Code PIX criado com sucesso")

	// Salva dados pendentes no Redis com chave baseada no pixId
	pendingData := gateway.PendingRegistrationData{
		Name:      userData.Name,
		Email:     userData.Email,
		Password:  string(hashedPassword),
		Tax:       userData.Tax,
		Cellphone: userData.Cellphone,
		PlanID:    plan.ID,
	}
	jsonData, err := json.Marshal(pendingData)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar dados pendentes")
		return nil, fmt.Errorf("erro ao serializar dados pendentes: %w", err)
	}

	redisKey := "pending_reg:" + pixData.ID
	ttl := 1 * time.Hour
	err = uc.redisClient.Set(ctx, redisKey, jsonData, ttl).Err()
	if err != nil {
		log.Error().Err(err).Str("redis_key", redisKey).Msg("Erro ao salvar dados pendentes no Redis")
		return nil, fmt.Errorf("erro ao salvar dados pendentes no Redis: %w", err)
	}
	log.Info().Str("redis_key", redisKey).Dur("ttl", ttl).Msg("Dados de registro pendente salvos no Redis")

	return &PixQRCodeResult{
		PixID:        pixData.ID,
		BrCode:       pixData.BrCode,
		BrCodeBase64: pixData.BrCodeBase64,
		ExpiresAt:    pixData.ExpiresAt,
	}, nil
}

// CheckPixStatus consulta o status de pagamento de um QR Code PIX.
// Se o status for PAID, enfileira a task de registro do usuário (idempotente).
func (uc *PaymentUsecase) CheckPixStatus(ctx context.Context, pixId string, asynqClient *asynq.Client) (string, error) {
	log := logging.Logger.With().Str("usecase", "PaymentUsecase").Str("method", "CheckPixStatus").Str("pix_id", pixId).Logger()

	// Valida que o pixId tem dados pendentes no Redis
	redisKey := "pending_reg:" + pixId
	exists, err := uc.redisClient.Exists(ctx, redisKey).Result()
	if err != nil {
		log.Error().Err(err).Msg("Erro ao verificar existência no Redis")
		return "", fmt.Errorf("erro ao verificar registro pendente: %w", err)
	}
	if exists == 0 {
		return "", fmt.Errorf("registro pendente não encontrado para pix_id: %s", pixId)
	}

	// Consulta status na AbacatePay
	status, err := uc.paymentGateway.CheckPixQRCodeStatus(ctx, pixId)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao consultar status do PIX na AbacatePay")
		return "", fmt.Errorf("erro ao consultar status do PIX: %w", err)
	}

	log.Info().Str("status", status).Msg("Status do PIX consultado")

	if status == "PAID" {
		// Idempotência: verifica se já enfileirou a task para este pixId
		paidFlag := "pix_paid:" + pixId
		wasSet, err := uc.redisClient.SetNX(ctx, paidFlag, "1", 1*time.Hour).Result()
		if err != nil {
			log.Error().Err(err).Msg("Erro ao setar flag de pagamento no Redis")
		}

		if wasSet {
			// Primeira vez detectando PAID — enfileira task
			log.Info().Msg("PIX pago — enfileirando task de registro")

			taskPayload, err := json.Marshal(tasks.CompleteRegistrationPayload{
				PendingRegistrationID: pixId,
				CustomerEmail:         pixId, // será resolvido no CompleteRegistration
			})
			if err != nil {
				log.Error().Err(err).Msg("Erro ao serializar payload da task")
				return status, nil // retorna PAID mesmo assim
			}

			task := asynq.NewTask(tasks.TypeCompleteRegistration, taskPayload, asynq.MaxRetry(5))
			info, enqErr := asynqClient.Enqueue(task, asynq.Queue("critical"))
			if enqErr != nil {
				log.Error().Err(enqErr).Msg("Erro ao enfileirar task de registro")
			} else {
				log.Info().Str("task_id", info.ID).Msg("Task de registro enfileirada com sucesso")
			}
		} else {
			log.Info().Msg("PIX já processado anteriormente (flag existente)")
		}
	}

	return status, nil
}

func (uc *PaymentUsecase) CompleteRegistration(ctx context.Context, pendingRegistrationID string) (*model.User, error) {
	log := logging.Logger.With().Str("usecase", "PaymentUsecase").Str("method", "CompleteRegistration").Str("pending_reg_id", pendingRegistrationID).Logger()

	redisKey := "pending_reg:" + pendingRegistrationID
	redisClient := uc.redisClient

	log.Info().Msg("Buscando dados de registro pendente no Redis")
	jsonData, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Warn().Msg("Registro pendente não encontrado no Redis (pode já ter sido processado ou expirado)")
			return nil, fmt.Errorf("registro pendente não encontrado ou expirado (ID: %s)", pendingRegistrationID)
		}
		log.Error().Err(err).Msg("Erro ao buscar dados do Redis")
		return nil, fmt.Errorf("erro ao buscar dados do Redis: %w", err)
	}
	log.Info().Msg("Dados de registro pendente encontrados no Redis")

	var pendingData gateway.PendingRegistrationData
	if err := json.Unmarshal([]byte(jsonData), &pendingData); err != nil {
		log.Error().Err(err).Msg("Erro ao decodificar dados pendentes do Redis")
		return nil, fmt.Errorf("erro ao decodificar dados pendentes: %w", err)
	}

	log.Info().Str("email", pendingData.Email).Msg("Tentando criar usuário no banco de dados")
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	userToCreate := model.User{
		Name:      pendingData.Name,
		Email:     pendingData.Email,
		Password:  pendingData.Password, // Already hashed before Redis storage
		Tax:       &pendingData.Tax,
		Cellphone: &pendingData.Cellphone,
		PlanID:    &pendingData.PlanID,
		ExpiresAt: &expiresAt,
	}

	newUser, err := uc.userUsecase.CreateUserWithHashedPassword(userToCreate)
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") || strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			log.Warn().Str("email", pendingData.Email).Msg("Usuário já existe, registro não duplicado.")
			existingUser, findErr := uc.userUsecase.GetUserByEmail(pendingData.Email)
			if findErr != nil {
				log.Error().Err(findErr).Str("email", pendingData.Email).Msg("Erro ao buscar usuário existente após erro de duplicidade.")
				return nil, findErr
			}

			// Renew subscription for existing user
			newExpiry := time.Now().Add(30 * 24 * time.Hour)
			if updateErr := uc.userUsecase.UpdateExpiresAt(existingUser.Id, newExpiry); updateErr != nil {
				log.Error().Err(updateErr).Int("user_id", existingUser.Id).Msg("Erro ao renovar assinatura do usuário")
			}

			if delErr := redisClient.Del(ctx, redisKey).Err(); delErr != nil {
				log.Error().Err(delErr).Str("redis_key", redisKey).Msg("Falha ao deletar chave do Redis após detectar usuário duplicado")
			}
			log.Info().Int("user_id", existingUser.Id).Msg("Registro considerado completo (usuário já existia)")
			return &existingUser, nil
		}
		log.Error().Err(err).Msg("Erro ao criar usuário no banco de dados")
		return nil, fmt.Errorf("erro ao criar usuário no banco de dados: %w", err)
	}
	log.Info().Int("user_id", newUser.Id).Msg("Usuário criado com sucesso no banco de dados")

	if err := redisClient.Del(ctx, redisKey).Err(); err != nil {
		log.Warn().Err(err).Str("redis_key", redisKey).Msg("Falha ao deletar chave do Redis após registro bem-sucedido")
	} else {
		log.Info().Str("redis_key", redisKey).Msg("Chave Redis de registro pendente deletada")
	}

	return &newUser, nil
}

// cleanNumericString remove todos os caracteres não numéricos de uma string (CPF, celular, etc.)
func cleanNumericString(s string) string {
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
