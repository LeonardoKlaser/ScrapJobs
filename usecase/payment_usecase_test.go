package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"web-scrapper/gateway"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupPaymentTest(t *testing.T) (*PaymentUsecase, *mocks.MockPaymentGateway, *mocks.MockPlanRepository, *mocks.MockUserRepository, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockGW := new(mocks.MockPaymentGateway)
	mockPlanRepo := new(mocks.MockPlanRepository)
	mockUserRepo := new(mocks.MockUserRepository)

	userUsecase := NewUserUsercase(mockUserRepo)
	uc := NewPaymentUsecase(mockGW, redisClient, userUsecase, mockPlanRepo)

	return uc, mockGW, mockPlanRepo, mockUserRepo, mr
}

func TestCreatePayment_Success(t *testing.T) {
	uc, mockGW, mockPlanRepo, _, _ := setupPaymentTest(t)
	ctx := context.Background()

	plan := &model.Plan{ID: 1, Name: "Pro", Price: 29.90}
	mockPlanRepo.On("GetPlanByID", 1).Return(plan, nil)
	mockGW.On("CreateBilling", ctx, plan, mock.AnythingOfType("*gateway.InitiatePaymentRequest")).
		Return("https://pay.example.com", "pending-123", nil)

	userData := gateway.InitiatePaymentRequest{
		Name:          "John",
		Email:         "john@test.com",
		Password:      "secret123",
		Tax:           "123.456.789-00",
		Cellphone:     "(11) 99999-9999",
		Methods:       []string{"PIX"},
		BillingPeriod: "monthly",
	}

	url, err := uc.CreatePayment(ctx, 1, userData)
	assert.NoError(t, err)
	assert.Equal(t, "https://pay.example.com", url)
	mockPlanRepo.AssertExpectations(t)
	mockGW.AssertExpectations(t)
}

func TestCreatePayment_PlanNotFound(t *testing.T) {
	uc, _, mockPlanRepo, _, _ := setupPaymentTest(t)
	ctx := context.Background()

	mockPlanRepo.On("GetPlanByID", 999).Return((*model.Plan)(nil), nil)

	userData := gateway.InitiatePaymentRequest{Name: "Test", Email: "t@t.com", Password: "123456"}
	_, err := uc.CreatePayment(ctx, 999, userData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não encontrado")
}

func TestCreatePayment_GatewayError(t *testing.T) {
	uc, mockGW, mockPlanRepo, _, _ := setupPaymentTest(t)
	ctx := context.Background()

	plan := &model.Plan{ID: 1, Name: "Pro", Price: 29.90}
	mockPlanRepo.On("GetPlanByID", 1).Return(plan, nil)
	mockGW.On("CreateBilling", ctx, plan, mock.AnythingOfType("*gateway.InitiatePaymentRequest")).
		Return("", "", fmt.Errorf("gateway timeout"))

	userData := gateway.InitiatePaymentRequest{
		Name:          "John",
		Email:         "john@test.com",
		Password:      "secret123",
		Tax:           "12345678900",
		Cellphone:     "11999999999",
		Methods:       []string{"PIX"},
		BillingPeriod: "monthly",
	}

	_, err := uc.CreatePayment(ctx, 1, userData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway")
}

func TestCompleteRegistration_Success(t *testing.T) {
	uc, _, _, mockUserRepo, mr := setupPaymentTest(t)
	ctx := context.Background()

	pendingData := gateway.PendingRegistrationData{
		Name:      "John",
		Email:     "john@test.com",
		Password:  "$2a$10$hashedpassword",
		Tax:       "12345678900",
		Cellphone: "11999999999",
		PlanID:    1,
	}
	jsonData, _ := json.Marshal(pendingData)
	mr.Set("pending_reg:abc-123", string(jsonData))

	// CreateUserWithHashedPassword first calls GetUserByEmail.
	// Return empty user with no error to indicate user does not exist yet.
	mockUserRepo.On("GetUserByEmail", "john@test.com").Return(model.User{}, nil)

	planID := 1
	tax := "12345678900"
	cellphone := "11999999999"
	expectedUser := model.User{
		Id:       1,
		Name:     "John",
		Email:    "john@test.com",
		Password: "$2a$10$hashedpassword",
		Tax:      &tax,
		Cellphone: &cellphone,
		PlanID:   &planID,
	}
	mockUserRepo.On("CreateUser", mock.AnythingOfType("model.User")).Return(expectedUser, nil)

	user, err := uc.CompleteRegistration(ctx, "abc-123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 1, user.Id)
	mockUserRepo.AssertExpectations(t)
}

func TestCompleteRegistration_NotFound(t *testing.T) {
	uc, _, _, _, _ := setupPaymentTest(t)
	ctx := context.Background()

	user, err := uc.CompleteRegistration(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "não encontrado")
}
