package interfaces

import (
	"context"
	"web-scrapper/gateway"
	"web-scrapper/model"
)

type PaymentGatewayInterface interface {
	CreateBilling(ctx context.Context, plan *model.Plan, userData *gateway.InitiatePaymentRequest) (paymentURL string, pendingRegistrationID string, err error)
	CreatePixQRCode(ctx context.Context, amountCents int, expiresInSeconds int, description string, customer *gateway.PixCustomer) (*gateway.PixQRCodeData, error)
	CheckPixQRCodeStatus(ctx context.Context, pixId string) (string, error)
}
