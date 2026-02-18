package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"web-scrapper/logging"
	"web-scrapper/model"

	"github.com/google/uuid"
)

type AbacatePayGateway struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewAbacatePayGateway() *AbacatePayGateway {
	apiKey := os.Getenv("ABACATEPAY_API_KEY")
	baseURL := os.Getenv("ABACATEPAY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.abacatepay.com"
		logging.Logger.Warn().Msg("ABACATEPAY_BASE_URL não definida, usando default de produção")
	}
	if apiKey == "" {
		logging.Logger.Error().Msg("ABACATEPAY_API_KEY não está definida")
	}

	return &AbacatePayGateway{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		apiKey:     apiKey,
		baseURL:    baseURL,
	}
}

// CreateBilling cria uma cobrança na AbacatePay.
// Retorna: (paymentURL, pendingRegistrationID, error)
func (a *AbacatePayGateway) CreateBilling(ctx context.Context, plan *model.Plan, userData *InitiatePaymentRequest) (string, string, error) {
	if a.apiKey == "" {
		return "", "", errors.New("ABACATEPAY_API_KEY não está definida")
	}

	// ID único que usamos para rastrear o registro pendente no Redis
	pendingRegistrationID := uuid.New().String()

	body := &CreateBillingBody{
		Frequency:     userData.Frequency,
		Methods:       userData.Methods,
		CompletionUrl: os.Getenv("FRONTEND_URL") + "/payment-confirmation",
		ReturnUrl:     os.Getenv("FRONTEND_URL") + "/checkout/" + fmt.Sprintf("%d", plan.ID),
		ExternalId:    pendingRegistrationID, // campo correto conforme doc AbacatePay
		Products: []*BillingProduct{
			{
				ExternalId:  fmt.Sprintf("plan-%d", plan.ID),
				Name:        plan.Name,
				Description: plan.Name,
				Quantity:    1,
				Price:       int(plan.Price * 100), // em centavos
			},
		},
		Customer: &BillingCustomer{
			Email:     userData.Email,
			Name:      userData.Name,
			TaxId:     userData.Tax,
			Cellphone: userData.Cellphone,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("erro ao converter payload para JSON: %w", err)
	}

	url := a.baseURL + "/v1/billing/create"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("erro ao executar requisição para AbacatePay: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logging.Logger.Error().Err(readErr).Int("status_code", resp.StatusCode).Msg("Erro ao ler corpo da resposta da AbacatePay")
		return "", "", fmt.Errorf("erro ao ler corpo da resposta da AbacatePay: %w", readErr)
	}

	logging.Logger.Debug().Int("status_code", resp.StatusCode).RawJSON("abacatepay_response", respBodyBytes).Msg("Resposta recebida da AbacatePay")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		if json.Unmarshal(respBodyBytes, &errResp) == nil {
			return "", "", fmt.Errorf("API da AbacatePay retornou status %d: %v", resp.StatusCode, errResp)
		}
		return "", "", fmt.Errorf("API da AbacatePay retornou status %d", resp.StatusCode)
	}

	var billingResponse CreateBillingResponse
	if err := json.Unmarshal(respBodyBytes, &billingResponse); err != nil {
		return "", "", fmt.Errorf("erro ao decodificar resposta da AbacatePay: %w", err)
	}

	if billingResponse.Data == nil {
		return "", "", fmt.Errorf("resposta da AbacatePay não contém dados de billing")
	}

	// A AbacatePay retorna a URL de pagamento no campo "url"
	paymentURL := billingResponse.Data.URL
	if paymentURL == "" {
		return "", "", fmt.Errorf("AbacatePay não retornou URL de pagamento")
	}

	return paymentURL, pendingRegistrationID, nil
}
