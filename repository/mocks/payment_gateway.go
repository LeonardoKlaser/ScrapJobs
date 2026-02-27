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
