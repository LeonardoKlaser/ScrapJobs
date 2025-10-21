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


func (a *AbacatePayGateway) CreateBilling(ctx context.Context, body *CreateBillingBody) (*CreateBillingResponse, error) {
	if a.apiKey == "" {
		return nil, errors.New("ABACATEPAY_API_KEY não está definida")
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter payload para JSON: %w", err)
	}

	url := a.baseURL + "/v1/billing/create"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar requisição para AbacatePay: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API da AbacatePay retornou status %d: %v", resp.StatusCode, errResp)
	}

	var billingResponse CreateBillingResponse
	if err := json.NewDecoder(resp.Body).Decode(&billingResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta da AbacatePay: %w", err)
	}

	return &billingResponse, nil
}