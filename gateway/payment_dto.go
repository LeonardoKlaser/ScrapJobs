package gateway


type InitiatePaymentRequest struct {
	// Dados do Usuário (sem senha aqui, trataremos no usecase)
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"` 
	Tax       string `json:"tax" binding:"required"`            
	Cellphone string `json:"cellphone" binding:"required"`     

	Methods   []string `json:"methods" binding:"required"`
	Frequency string   `json:"frequency" binding:"required"`
}

// PendingRegistrationData: Dados armazenados temporariamente (Redis)
type PendingRegistrationData struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	HashedPassword  string `json:"hashed_password"` 
	Tax             string `json:"tax"`
	Cellphone       string `json:"cellphone"`
	PlanID          int    `json:"plan_id"`
}

type BillingProduct struct {
	ExternalId  string `json:"externalId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"` 
}


type BillingCustomer struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Cellphone  string `json:"cellphone"`
	TaxId  string `json:"taxId"`
}


type CreateBillingBody struct {
	Frequency     string           `json:"frequency"` 
	Methods       []string         `json:"methods"`   
	CompletionUrl string           `json:"completionUrl"`
	ReturnUrl     string           `json:"returnUrl"`
	Products      []*BillingProduct `json:"products"`
	Customer      *BillingCustomer `json:"customer"`

	ExternalReference string `json:"external_reference,omitempty"` 
	Metadata          map[string]string `json:"metadata,omitempty"`
}


type CreateBillingResponseData struct {
	URL string `json:"url"`
}


type CreateBillingResponse struct {
	Data *CreateBillingResponseData `json:"data"`
}


type WebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		Object struct {
			// Campos que a AbacatePay retorna no webhook - ESSENCIAL
			ExternalReference string            `json:"external_reference"` 
			Metadata          map[string]string `json:"metadata"`
			Status            string            `json:"status"`
			Customer          struct { // Pode ser útil para dupla verificação
				Email string `json:"email"`
			} `json:"customer"`
			
		} `json:"object"`
	} `json:"data"`
}