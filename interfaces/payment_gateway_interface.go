package interfaces

import (
	"context"
	"web-scrapper/gateway"
	"web-scrapper/model"
)

type PaymentGatewayInterface interface {
	CreateBilling(ctx context.Context, plan *model.Plan, userData *gateway.InitiatePaymentRequest) (paymentURL string, pendingRegistrationID string, err error)
}
