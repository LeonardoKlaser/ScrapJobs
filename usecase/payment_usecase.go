package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"web-scrapper/gateway"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type PaymentUsecase struct {
	paymentGateway *gateway.AbacatePayGateway
	redisOpt       asynq.RedisConnOpt
	userUsecase    *UserUsecase
	planRepository interfaces.PlanRepositoryInterface
}

func NewPaymentUsecase(gw *gateway.AbacatePayGateway, redisOpt asynq.RedisConnOpt, userUsecase *UserUsecase, planRp interfaces.PlanRepositoryInterface) *PaymentUsecase {
	return &PaymentUsecase{paymentGateway: gw, redisOpt: redisOpt, userUsecase: userUsecase, planRepository: planRp}
}

func (uc *PaymentUsecase) getRedisClient() redis.UniversalClient {
	return uc.redisOpt.MakeRedisClient().(redis.UniversalClient)
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, planID int, userData gateway.InitiatePaymentRequest) (string, error) {
	log := logging.Logger.With().Str("usecase", "PaymentUsecase").Str("method", "InitiatePayment").Logger()

	plan, err := uc.planRepository.GetPlanByID(planID)
	if err != nil || plan == nil {
		if err == nil && plan == nil {
			err = errors.New("plano não encontrado")
		}
		log.Error().Err(err).Int("plan_id", planID).Msg("Erro ao buscar plano")
		return "", fmt.Errorf("plano com ID %d não encontrado: %w", planID, err)
	}
	log.Info().Int("plan_id", planID).Str("plan_name", plan.Name).Msg("Plano encontrado")

	// Limpa CPF e celular antes de armazenar
	userData.Tax = cleanNumericString(userData.Tax)
	userData.Cellphone = cleanNumericString(userData.Cellphone)

	// Hash password before storing in Redis to avoid plaintext exposure
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao gerar hash da senha")
		return "", fmt.Errorf("erro ao processar senha: %w", err)
	}

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
		return "", fmt.Errorf("erro ao serializar dados pendentes: %w", err)
	}

	log.Info().Str("email", userData.Email).Int("plan_id", planID).Msg("Iniciando cobrança na AbacatePay")
	paymentURL, pendingRegistrationID, err := uc.paymentGateway.CreateBilling(ctx, plan, &userData)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao iniciar cobrança na gateway AbacatePay")
		return "", fmt.Errorf("erro ao iniciar cobrança na gateway: %w", err)
	}
	log.Info().Str("pending_reg_id", pendingRegistrationID).Msg("Cobrança iniciada com sucesso")

	redisKey := "pending_reg:" + pendingRegistrationID
	redisClient := uc.getRedisClient()
	defer redisClient.Close()

	ttl := 1 * time.Hour
	err = redisClient.Set(ctx, redisKey, jsonData, ttl).Err()
	if err != nil {
		log.Error().Err(err).Str("redis_key", redisKey).Msg("Erro ao salvar dados pendentes no Redis")
		return "", fmt.Errorf("erro ao salvar dados pendentes no Redis: %w", err)
	}
	log.Info().Str("redis_key", redisKey).Dur("ttl", ttl).Msg("Dados de registro pendente salvos no Redis")

	return paymentURL, nil
}

func (uc *PaymentUsecase) CompleteRegistration(ctx context.Context, pendingRegistrationID string) (*model.User, error) {
	log := logging.Logger.With().Str("usecase", "PaymentUsecase").Str("method", "CompleteRegistration").Str("pending_reg_id", pendingRegistrationID).Logger()

	redisKey := "pending_reg:" + pendingRegistrationID
	redisClient := uc.getRedisClient()
	defer redisClient.Close()

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
	userToCreate := model.User{
		Name:      pendingData.Name,
		Email:     pendingData.Email,
		Password:  pendingData.Password, // Already hashed before Redis storage
		Tax:       &pendingData.Tax,
		Cellphone: &pendingData.Cellphone,
		PlanID:    &pendingData.PlanID,
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
