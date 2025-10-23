package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"web-scrapper/gateway"
	"web-scrapper/model"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type PaymentUsecase struct {
	paymentGateway *gateway.AbacatePayGateway
	redisClient *asynq.Client
	userUsecase *UserUsecase
}

func NewPaymentUsecase(gw *gateway.AbacatePayGateway, redisClient *asynq.Client, userUsecase *UserUsecase) *PaymentUsecase {
	return &PaymentUsecase{paymentGateway: gw, redisClient: redisClient, userUsecase: userUsecase}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, plan model.Plan, userData gateway.InitiatePaymentRequest) (string, error) {

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	pendingData := gateway.PendingRegistrationData{
		Name:            userData.Name,
		Email:           userData.Email,
		HashedPassword:  string(hashedPasswordBytes), 
		Tax:             userData.Tax,                
		Cellphone:       userData.Cellphone,          
		PlanID:          plan.ID,
	}

	jsonData, err := json.Marshal(pendingData)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar dados pendentes: %w", err)
	}

	// 4. Chamar a gateway para criar a cobrança e obter nosso ID temporário
	billingResponse, pendingRegistrationID, err := uc.paymentGateway.CreateBilling(ctx, &plan, &userData)
	if err != nil {
		return "", fmt.Errorf("erro ao iniciar cobrança na gateway: %w", err)
	}

	if billingResponse.Data == nil || billingResponse.Data.URL == "" {
		return "", errors.New("resposta da AbacatePay inválida")
	}

	err = uc.redisClient.Set(ctx, "pending_reg:"+pendingRegistrationID, jsonData, 1*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("erro ao salvar dados pendentes no Redis: %w", err)
	}

	return billingResponse.Data.URL, nil
}


func (uc *PaymentUsecase) CompleteRegistration(ctx context.Context, pendingRegistrationID string) (*model.User, error) {
	jsonData, err := uc.redisClient.Get(ctx, "pending_reg:"+pendingRegistrationID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("registro pendente não encontrado ou expirado (ID: %s)", pendingRegistrationID)
		}
		return nil, fmt.Errorf("erro ao buscar dados do Redis: %w", err)
	}

	
	var pendingData gateway.PendingRegistrationData
	if err := json.Unmarshal([]byte(jsonData), &pendingData); err != nil {
		return nil, fmt.Errorf("erro ao decodificar dados pendentes: %w", err)
	}

	userToCreate := model.User{
		Name:     pendingData.Name,
		Email:    pendingData.Email,
		Password: pendingData.HashedPassword,
		Tax:      &pendingData.Tax,
		Cellphone:&pendingData.Cellphone,
		Plan:     &model.Plan{ID: pendingData.PlanID}, 
	}

	err = uc.userUsecase.CreateUser(userToCreate);
	if err != nil {
	    return nil, fmt.Errorf("erro ao criar usuário no banco de dados: %w", err)
	}

	if err := uc.redisClient.Del(ctx, "pending_reg:"+pendingRegistrationID).Err(); err != nil {
		fmt.Printf("AVISO: Falha ao deletar chave do Redis após registro: %v\n", err)
	}

	return &userToCreate, nil
}