package gateway

// InitiatePaymentRequest: corpo da requisição do frontend para iniciar pagamento
type InitiatePaymentRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Tax       string `json:"tax" binding:"required"`
	Cellphone string `json:"cellphone" binding:"required"`

	Methods       []string `json:"methods" binding:"required"`
	BillingPeriod string   `json:"billing_period" binding:"required,oneof=monthly quarterly"`
}

// PendingRegistrationData: Dados armazenados temporariamente no Redis até o pagamento ser confirmado
type PendingRegistrationData struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"` // senha já com hash bcrypt (hash feito no CreatePayment antes de salvar no Redis)
	Tax       string `json:"tax"`
	Cellphone string `json:"cellphone"`
	PlanID    int    `json:"plan_id"`
	PixID     string `json:"pix_id,omitempty"` // ID do QR Code PIX para cleanup de chaves Redis
}

// BillingProduct: produto enviado para a AbacatePay
type BillingProduct struct {
	ExternalId  string `json:"externalId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"` // em centavos
}

// BillingCustomer: dados do cliente enviados para a AbacatePay
type BillingCustomer struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Cellphone string `json:"cellphone"`
	TaxId     string `json:"taxId"`
}

// CreateBillingBody: corpo da requisição POST /v1/billing/create
type CreateBillingBody struct {
	Frequency     string            `json:"frequency"`
	Methods       []string          `json:"methods"`
	CompletionUrl string            `json:"completionUrl"`
	ReturnUrl     string            `json:"returnUrl"`
	Products      []*BillingProduct `json:"products"`
	Customer      *BillingCustomer  `json:"customer"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// CreateBillingResponseData: dados retornados pela AbacatePay ao criar billing
type CreateBillingResponseData struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

// CreateBillingResponse: resposta completa da AbacatePay
type CreateBillingResponse struct {
	Data  *CreateBillingResponseData `json:"data"`
	Error interface{}                `json:"error"`
}

// --- Webhook Payloads (baseado na documentação oficial AbacatePay) ---

// WebhookBillingCustomer: cliente dentro do payload do webhook
type WebhookBillingCustomer struct {
	ID       string            `json:"id"`
	Metadata map[string]string `json:"metadata"` // contém email, name, cellphone, taxId
}

// WebhookBillingProduct: produto dentro do payload do webhook
type WebhookBillingProduct struct {
	ID         string `json:"id"`
	ExternalId string `json:"externalId"`
	Quantity   int    `json:"quantity"`
}

// WebhookBilling: objeto billing dentro do payload do webhook billing.paid
type WebhookBilling struct {
	ID         string                  `json:"id"`
	Amount     int                     `json:"amount"`
	PaidAmount int                     `json:"paidAmount"`
	Status     string                  `json:"status"` // "PAID"
	Frequency  string                  `json:"frequency"`
	Kind       []string                `json:"kind"`
	Customer   *WebhookBillingCustomer `json:"customer"`
	Products   []WebhookBillingProduct `json:"products"`
}

// WebhookPayment: dados do pagamento dentro do webhook
type WebhookPayment struct {
	Amount int    `json:"amount"`
	Fee    int    `json:"fee"`
	Method string `json:"method"`
}

// WebhookData: campo "data" do payload do webhook
type WebhookData struct {
	Payment *WebhookPayment `json:"payment,omitempty"`
	Billing *WebhookBilling `json:"billing,omitempty"`
}

// WebhookPayload: payload completo recebido no webhook da AbacatePay
type WebhookPayload struct {
	ID      string      `json:"id"`
	Event   string      `json:"event"` // "billing.paid", "withdraw.done", etc.
	Data    WebhookData `json:"data"`
	DevMode bool        `json:"devMode"`
}

// --- PIX QR Code Payloads (POST /v1/pixQrCode/create) ---

// PixCustomer: dados do cliente para criação do QR Code PIX
type PixCustomer struct {
	Name      string `json:"name"`
	Cellphone string `json:"cellphone"`
	Email     string `json:"email"`
	TaxId     string `json:"taxId"`
}

// CreatePixQRCodeBody: corpo da requisição POST /v1/pixQrCode/create
type CreatePixQRCodeBody struct {
	Amount      int                    `json:"amount"`
	ExpiresIn   int                    `json:"expiresIn"`
	Description string                 `json:"description,omitempty"`
	Customer    *PixCustomer           `json:"customer,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PixQRCodeData: dados do QR Code retornados pela AbacatePay
type PixQRCodeData struct {
	ID           string `json:"id"`
	Amount       int    `json:"amount"`
	Status       string `json:"status"`
	BrCode       string `json:"brCode"`
	BrCodeBase64 string `json:"brCodeBase64"`
	ExpiresAt    string `json:"expiresAt"`
}

// CreatePixQRCodeResponse: resposta completa da AbacatePay ao criar QR Code PIX
type CreatePixQRCodeResponse struct {
	Data  *PixQRCodeData `json:"data"`
	Error interface{}    `json:"error"`
}

// CheckPixStatusResponseData: dados retornados pela AbacatePay ao checar status do QR Code
type CheckPixStatusResponseData struct {
	Status    string `json:"status"`
	ExpiresAt string `json:"expiresAt"`
}

// CheckPixStatusResponse: resposta completa da AbacatePay ao checar status
type CheckPixStatusResponse struct {
	Data  *CheckPixStatusResponseData `json:"data"`
	Error interface{}                 `json:"error"`
}
