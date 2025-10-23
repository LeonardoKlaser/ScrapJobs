package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	"web-scrapper/model"

	"github.com/google/uuid"
)

type AbacatePayGateway struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewAbacatePayGateway() *AbacatePayGateway {
	return &AbacatePayGateway{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		apiKey:     os.Getenv("ABACATEPAY_API_KEY"),
		baseURL:    "https://api.abacatepay.com", 
	}
}


func (a *AbacatePayGateway) CreateBilling(ctx context.Context,plan *model.Plan ,userData *InitiatePaymentRequest) (*CreateBillingResponse, string ,error) {
	if a.apiKey == "" {
		return nil, "", errors.New("ABACATEPAY_API_KEY não está definida")
	}

	pendingRegistrationID := uuid.New().String()

	body := &CreateBillingBody{
		Frequency:     userData.Frequency,
		Methods:       userData.Methods,
		CompletionUrl: "http://localhost:5173/payment-confirmation", // Pode adicionar o ID aqui se quiser: ?regId=" + pendingRegistrationID
		ReturnUrl:     "http://localhost:5173/checkout",
		Products: []*BillingProduct{
			{
				ExternalId:  fmt.Sprintf("plan-%d", plan.ID), // ID do produto/plano
				Name:        plan.Name,
				Description: plan.Name,
				Quantity:    1,
				Price:       int(plan.Price * 100),
			},
		},
		Customer: &BillingCustomer{
			Email: userData.Email,
			Name:  userData.Name,
			TaxId: userData.Tax,
			Cellphone: userData.Cellphone,
		},
		ExternalReference: pendingRegistrationID, // Enviando nosso ID único
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao converter payload para JSON: %w", err)
	}

	url := a.baseURL + "/v1/billing/create"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, "", fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao executar requisição para AbacatePay: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, "", fmt.Errorf("API da AbacatePay retornou status %d: %v", resp.StatusCode, errResp)
	}

	var billingResponse CreateBillingResponse
	if err := json.NewDecoder(resp.Body).Decode(&billingResponse); err != nil {
		return nil, "", fmt.Errorf("erro ao decodificar resposta da AbacatePay: %w", err)
	}

	return &billingResponse, pendingRegistrationID, nil
}