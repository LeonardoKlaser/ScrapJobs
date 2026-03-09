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

	pixData := &gateway.PixQRCodeData{
		ID:           "pix-123",
		Amount:       2990,
		Status:       "PENDING",
		BrCode:       "00020126...",
		BrCodeBase64: "iVBORw0KGgo...",
		ExpiresAt:    "2026-03-08T23:00:00Z",
	}
	mockGW.On("CreatePixQRCode", ctx, 2990, 900, "ScrapJobs - Pro (Mensal)", mock.AnythingOfType("*gateway.PixCustomer")).
		Return(pixData, nil)

	userData := gateway.InitiatePaymentRequest{
		Name:          "John",
		Email:         "john@test.com",
		Password:      "secret123",
		Tax:           "123.456.789-00",
		Cellphone:     "(11) 99999-9999",
		Methods:       []string{"PIX"},
		BillingPeriod: "monthly",
	}

	result, err := uc.CreatePayment(ctx, 1, userData)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "pix-123", result.PixID)
	assert.Equal(t, "00020126...", result.BrCode)
	assert.Equal(t, "iVBORw0KGgo...", result.BrCodeBase64)
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
	mockGW.On("CreatePixQRCode", ctx, 2990, 900, "ScrapJobs - Pro (Mensal)", mock.AnythingOfType("*gateway.PixCustomer")).
		Return((*gateway.PixQRCodeData)(nil), fmt.Errorf("gateway timeout"))

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
	assert.Contains(t, err.Error(), "PIX")
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
