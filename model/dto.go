package model

// --- Generic Responses ---

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error" example:"mensagem de erro"`
}

// MessageResponse represents a success message response.
type MessageResponse struct {
	Message string `json:"message" example:"operação realizada com sucesso"`
}

// StatusResponse represents a status response.
type StatusResponse struct {
	Status string `json:"status" example:"received"`
}

// --- Auth ---

// LoginRequest represents login credentials.
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"senha123"`
}

// SignUpRequest represents user registration data.
type SignUpRequest struct {
	UserName  string  `json:"user_name" binding:"required" example:"João Silva"`
	Email     string  `json:"email" binding:"required" example:"user@example.com"`
	Password  string  `json:"password" binding:"required,min=6" example:"senha123"`
	Tax       *string `json:"tax,omitempty" example:"123.456.789-00"`
	Cellphone *string `json:"cellphone,omitempty" example:"11999998888"`
}

// --- User ---

// UpdateProfileRequest represents profile update data.
type UpdateProfileRequest struct {
	UserName  string  `json:"user_name" binding:"required" example:"João Silva"`
	Cellphone *string `json:"cellphone,omitempty" example:"11999998888"`
	Tax       *string `json:"tax,omitempty" example:"123.456.789-00"`
}

// ChangePasswordRequest represents password change data.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"senhaAntiga123"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"senhaNova456"`
}

// ValidateCheckoutRequest represents checkout validation input.
type ValidateCheckoutRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
	Tax   string `json:"tax" binding:"required" example:"123.456.789-00"`
}

// ValidateCheckoutResponse represents checkout validation result.
type ValidateCheckoutResponse struct {
	EmailExists bool `json:"email_exists" example:"false"`
	TaxExists   bool `json:"tax_exists" example:"false"`
}

// --- UserSite ---

// UpdateUserSiteFiltersRequest represents target words update.
type UpdateUserSiteFiltersRequest struct {
	TargetWords []string `json:"target_words"`
}

// --- Site Career ---

// SiteDTO represents a site with subscription status.
type SiteDTO struct {
	SiteName     string  `json:"site_name" example:"iFood"`
	BaseURL      string  `json:"base_url" example:"https://carreiras.ifood.com.br"`
	SiteID       int     `json:"site_id" example:"1"`
	LogoURL      *string `json:"logo_url,omitempty" example:"https://s3.amazonaws.com/logos/ifood.webp"`
	IsSubscribed bool    `json:"is_subscribed" example:"true"`
}

// SandboxScrapeResponse represents sandbox scrape result.
type SandboxScrapeResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Scraping concluído com sucesso"`
	Data    []Job  `json:"data"`
}

// SandboxScrapeErrorResponse represents sandbox scrape error.
type SandboxScrapeErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"falha ao executar scraping"`
	Message string `json:"message" example:"erro no scraping"`
}

// --- Payment ---

// CreatePaymentResponse represents payment creation result.
type CreatePaymentResponse struct {
	URL string `json:"url" example:"https://abacatepay.com/billing/abc123"`
}

// --- Health ---

// HealthResponse represents liveness check.
type HealthResponse struct {
	Status string `json:"status" example:"UP"`
}

// ReadinessResponse represents readiness check.
type ReadinessResponse struct {
	Database string `json:"database" example:"UP"`
	Redis    string `json:"redis" example:"UP"`
}

// --- Analysis ---

// AnalyzeJobRequest represents manual AI analysis request.
type AnalyzeJobRequest struct {
	JobID int `json:"job_id" binding:"required" example:"42"`
}

// SendAnalysisEmailRequest represents email sending request with analysis.
type SendAnalysisEmailRequest struct {
	JobID    int            `json:"job_id" binding:"required" example:"42"`
	Analysis ResumeAnalysis `json:"analysis" binding:"required"`
}

// --- Requested Site ---

// RequestedSiteRequest represents a user-submitted career page URL.
type RequestedSiteRequest struct {
	URL string `json:"url" binding:"required" example:"https://careers.google.com"`
}
