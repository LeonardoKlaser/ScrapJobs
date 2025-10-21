package usecase

import (
	"context"
	"errors"
	"fmt"
	"web-scrapper/gateway"
	"web-scrapper/model"
)

type PaymentUsecase struct {
	paymentGateway *gateway.AbacatePayGateway
}

func NewPaymentUsecase(gw *gateway.AbacatePayGateway) *PaymentUsecase {
	return &PaymentUsecase{paymentGateway: gw}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, plan model.Plan, user model.User, methods []string, frequency string) (string, error) {
	
	body := &gateway.CreateBillingBody{
		Frequency:     frequency, 
		Methods:       methods,
		CompletionUrl: "http://localhost:5173/payment-confirmation", 
		ReturnUrl:     "http://localhost:5173/checkout",           
		Products: []*gateway.BillingProduct{
			{
				ExternalId:  fmt.Sprintf("%d", plan.ID),
				Name:        plan.Name,
				Description: plan.Name,
				Quantity:    1,
				Price:       int(plan.Price * 100), 
			},
		},
		Customer: &gateway.BillingCustomer{
			Email: user.Email,
			Name:  user.Name,
			Cellphone: "(11) 4002-8922",
			TaxId: "034.692.480-44",
		},
	}

	
	resp, err := uc.paymentGateway.CreateBilling(ctx, body)
	if err != nil {
		return "", err
	}

	if resp.Data == nil || resp.Data.URL == "" {
		return "", errors.New("resposta da AbacatePay não contém URL de pagamento")
	}

	return resp.Data.URL, nil
}