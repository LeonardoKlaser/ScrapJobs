package gateway


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
			// Adicione os campos que você precisa para identificar o usuário/pagamento
			// Ex: CustomerEmail string `json:"customer_email"`
			// Ex: TransactionID string `json:"id"`
		} `json:"object"`
	} `json:"data"`
}