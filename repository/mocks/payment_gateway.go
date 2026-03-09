package mocks

import (
	"context"
	"web-scrapper/gateway"
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockPaymentGateway struct {
	mock.Mock
}

func (m *MockPaymentGateway) CreateBilling(ctx context.Context, plan *model.Plan, userData *gateway.InitiatePaymentRequest) (string, string, error) {
	args := m.Called(ctx, plan, userData)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockPaymentGateway) CreatePixQRCode(ctx context.Context, amountCents int, expiresInSeconds int, description string, customer *gateway.PixCustomer) (*gateway.PixQRCodeData, error) {
	args := m.Called(ctx, amountCents, expiresInSeconds, description, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gateway.PixQRCodeData), args.Error(1)
}

func (m *MockPaymentGateway) CheckPixQRCodeStatus(ctx context.Context, pixId string) (string, error) {
	args := m.Called(ctx, pixId)
	return args.String(0), args.Error(1)
}
