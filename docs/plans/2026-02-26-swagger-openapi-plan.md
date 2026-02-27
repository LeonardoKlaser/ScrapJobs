# Swagger/OpenAPI Documentation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document all 25+ REST endpoints with Swagger annotations, create typed DTOs, and serve Swagger UI in dev mode.

**Architecture:** Use swaggo/swag to generate OpenAPI 2.0 spec from Go comments. Add Swagger annotations to every controller handler. Create typed DTOs in `model/dto.go` to replace `gin.H{}` responses in documentation. Serve Swagger UI at `/swagger/*any` only when `GIN_MODE != "release"`.

**Tech Stack:** swaggo/swag, swaggo/gin-swagger, swaggo/files, Gin

---

### Task 1: Install swaggo dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Install swag CLI**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go install github.com/swaggo/swag/cmd/swag@latest
```

**Step 2: Add gin-swagger and swag files dependencies**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
go get github.com/swaggo/swag
```

**Step 3: Verify installation**

```bash
swag --version
```

Expected: prints version number

**Step 4: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add go.mod go.sum
git commit -m "chore: add swaggo/swag dependencies for API documentation"
```

---

### Task 2: Create DTOs for all API responses

**Files:**
- Create: `model/dto.go`

**Step 1: Create the DTO file with all typed request/response structs**

```go
package model

// --- Generic Responses ---

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"mensagem de erro"`
}

// MessageResponse represents a success message response
type MessageResponse struct {
	Message string `json:"message" example:"operação realizada com sucesso"`
}

// StatusResponse represents a status response
type StatusResponse struct {
	Status string `json:"status" example:"received"`
}

// --- Auth ---

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"senha123"`
}

// SignUpRequest represents user registration data
type SignUpRequest struct {
	UserName  string  `json:"user_name" binding:"required" example:"João Silva"`
	Email     string  `json:"email" binding:"required" example:"user@example.com"`
	Password  string  `json:"password" binding:"required,min=6" example:"senha123"`
	Tax       *string `json:"tax,omitempty" example:"123.456.789-00"`
	Cellphone *string `json:"cellphone,omitempty" example:"11999998888"`
}

// --- User ---

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	UserName  string  `json:"user_name" binding:"required" example:"João Silva"`
	Cellphone *string `json:"cellphone,omitempty" example:"11999998888"`
	Tax       *string `json:"tax,omitempty" example:"123.456.789-00"`
}

// ChangePasswordRequest represents password change data
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"senhaAntiga123"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"senhaNova456"`
}

// ValidateCheckoutRequest represents checkout validation input
type ValidateCheckoutRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
	Tax   string `json:"tax" binding:"required" example:"123.456.789-00"`
}

// ValidateCheckoutResponse represents checkout validation result
type ValidateCheckoutResponse struct {
	EmailExists bool `json:"email_exists" example:"false"`
	TaxExists   bool `json:"tax_exists" example:"false"`
}

// --- UserSite ---

// UpdateUserSiteFiltersRequest represents target words update
type UpdateUserSiteFiltersRequest struct {
	TargetWords []string `json:"target_words" example:"golang,backend,remoto"`
}

// --- Site Career ---

// SiteDTO represents a site with subscription status
type SiteDTO struct {
	SiteName     string  `json:"site_name" example:"iFood"`
	BaseURL      string  `json:"base_url" example:"https://carreiras.ifood.com.br"`
	SiteID       int     `json:"site_id" example:"1"`
	LogoURL      *string `json:"logo_url,omitempty" example:"https://s3.amazonaws.com/logos/ifood.webp"`
	IsSubscribed bool    `json:"is_subscribed" example:"true"`
}

// SandboxScrapeResponse represents sandbox scrape result
type SandboxScrapeResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Scraping concluído com sucesso"`
	Data    []Job  `json:"data"`
}

// SandboxScrapeErrorResponse represents sandbox scrape error
type SandboxScrapeErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"falha ao executar scraping"`
	Message string `json:"message" example:"erro no scraping"`
}

// --- Payment ---

// CreatePaymentResponse represents payment creation result
type CreatePaymentResponse struct {
	URL string `json:"url" example:"https://abacatepay.com/billing/abc123"`
}

// --- Health ---

// HealthResponse represents liveness check
type HealthResponse struct {
	Status string `json:"status" example:"UP"`
}

// ReadinessResponse represents readiness check
type ReadinessResponse struct {
	Database string `json:"database" example:"UP"`
	Redis    string `json:"redis" example:"UP"`
}

// --- Analysis ---

// AnalyzeJobRequest represents manual AI analysis request
type AnalyzeJobRequest struct {
	JobID int `json:"job_id" binding:"required" example:"42"`
}

// SendAnalysisEmailRequest represents email sending request
type SendAnalysisEmailRequest struct {
	JobID    int            `json:"job_id" binding:"required" example:"42"`
	Analysis ResumeAnalysis `json:"analysis" binding:"required"`
}

// --- Requested Site ---

// RequestedSiteRequest represents a user-submitted career page URL
type RequestedSiteRequest struct {
	URL string `json:"url" binding:"required" example:"https://careers.google.com"`
}
```

**Step 2: Verify compilation**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go build ./...
```

Expected: no errors

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add model/dto.go
git commit -m "feat: add typed DTOs for Swagger API documentation"
```

---

### Task 3: Add Swagger general annotations to main.go

**Files:**
- Modify: `cmd/api/main.go`

**Step 1: Add Swagger general info comments before `func main()`**

Add these comments right before the `func main()` line in `cmd/api/main.go`:

```go
// @title ScrapJobs API
// @version 1.0
// @description API para plataforma de scraping e matching de vagas de emprego.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name Authorization
func main() {
```

**Step 2: Add Swagger route (dev only) and imports**

Add to the imports section:
```go
_ "web-scrapper/docs/swagger"
ginSwagger "github.com/swaggo/gin-swagger"
swaggerFiles "github.com/swaggo/files"
```

Add the swagger route right before the `srv := &http.Server{` line:
```go
// Swagger documentation (dev only)
if os.Getenv("GIN_MODE") != "release" {
    server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

**Step 3: Verify compilation (will fail until `swag init` is run — that's ok)**

Note: This step can't compile yet because `docs/swagger` package doesn't exist. We'll generate it in Task 5.

**Step 4: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add cmd/api/main.go
git commit -m "feat: add Swagger general annotations and route to main.go"
```

---

### Task 4: Add Swagger annotations to all controllers

**Files:**
- Modify: `controller/user_controller.go`
- Modify: `controller/checkAuthController.go`
- Modify: `controller/curriculum_controller.go`
- Modify: `controller/user_site_controller.go`
- Modify: `controller/plan_controller.go`
- Modify: `controller/site_career_controller.go`
- Modify: `controller/payment_controller.go`
- Modify: `controller/dashboardDataController.go`
- Modify: `controller/health_controller.go`
- Modify: `controller/notification_controller.go`
- Modify: `controller/analysis_controller.go`
- Modify: `controller/requested_site_controller.go`
- Modify: `controller/admin_dashboard_controller.go`

Add Swagger comment blocks above each handler method. Below are ALL annotations needed:

#### 4.1 — user_controller.go

**SignIn:**
```go
// SignIn godoc
// @Summary Login do usuario
// @Description Autentica o usuario e retorna JWT em cookie HTTP-only
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Credenciais de login"
// @Success 200 {object} object "Cookie Authorization definido"
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /login [post]
```

**SignUp:**
```go
// SignUp godoc
// @Summary Registro de usuario
// @Description Cria uma nova conta de usuario
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.SignUpRequest true "Dados de registro"
// @Success 201 "Usuario criado com sucesso"
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /register [post]
```

**Logout:**
```go
// Logout godoc
// @Summary Logout do usuario
// @Description Remove o cookie de autenticacao
// @Tags Auth
// @Produce json
// @Success 200 {object} model.MessageResponse
// @Security CookieAuth
// @Router /api/logout [post]
```

**UpdateProfile:**
```go
// UpdateProfile godoc
// @Summary Atualizar perfil
// @Description Atualiza nome, telefone e CPF do usuario
// @Tags User
// @Accept json
// @Produce json
// @Param body body model.UpdateProfileRequest true "Dados do perfil"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/user/profile [patch]
```

**ChangePassword:**
```go
// ChangePassword godoc
// @Summary Alterar senha
// @Description Altera a senha do usuario autenticado
// @Tags User
// @Accept json
// @Produce json
// @Param body body model.ChangePasswordRequest true "Senhas antiga e nova"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/user/change-password [post]
```

**ValidateCheckout:**
```go
// ValidateCheckout godoc
// @Summary Validar dados de checkout
// @Description Verifica se email e CPF ja estao cadastrados
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.ValidateCheckoutRequest true "Email e CPF"
// @Success 200 {object} model.ValidateCheckoutResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/users/validate-checkout [post]
```

#### 4.2 — checkAuthController.go

**CheckAuthUser:**
```go
// CheckAuthUser godoc
// @Summary Verificar autenticacao
// @Description Retorna dados do usuario autenticado
// @Tags Auth
// @Produce json
// @Success 200 {object} model.User
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/me [get]
```

#### 4.3 — curriculum_controller.go

**CreateCurriculum:**
```go
// CreateCurriculum godoc
// @Summary Criar curriculo
// @Description Cria um novo curriculo para o usuario autenticado
// @Tags Curriculum
// @Accept json
// @Produce json
// @Param body body model.Curriculum true "Dados do curriculo"
// @Success 201 {object} model.Curriculum
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum [post]
```

**GetCurriculumByUserId:**
```go
// GetCurriculumByUserId godoc
// @Summary Listar curriculos
// @Description Retorna todos os curriculos do usuario autenticado
// @Tags Curriculum
// @Produce json
// @Success 200 {array} model.Curriculum
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum [get]
```

**UpdateCurriculum:**
```go
// UpdateCurriculum godoc
// @Summary Atualizar curriculo
// @Description Atualiza um curriculo existente
// @Tags Curriculum
// @Accept json
// @Produce json
// @Param id path int true "ID do curriculo"
// @Param body body model.Curriculum true "Dados atualizados"
// @Success 200 {object} model.Curriculum
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum/{id} [put]
```

**SetActiveCurriculum:**
```go
// SetActiveCurriculum godoc
// @Summary Ativar curriculo
// @Description Define um curriculo como ativo para o usuario
// @Tags Curriculum
// @Produce json
// @Param id path int true "ID do curriculo"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /curriculum/{id}/active [patch]
```

#### 4.4 — user_site_controller.go

**InsertUserSite:**
```go
// InsertUserSite godoc
// @Summary Inscrever-se em site
// @Description Adiciona inscricao do usuario em um site de carreiras
// @Tags UserSite
// @Accept json
// @Produce json
// @Param body body model.UserSiteRequest true "ID do site e palavras-chave"
// @Success 201 {object} object
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /userSite [post]
```

**DeleteUserSite:**
```go
// DeleteUserSite godoc
// @Summary Cancelar inscricao em site
// @Description Remove inscricao do usuario em um site de carreiras
// @Tags UserSite
// @Produce json
// @Param siteId path string true "ID do site"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /userSite/{siteId} [delete]
```

**UpdateUserSiteFilters:**
```go
// UpdateUserSiteFilters godoc
// @Summary Atualizar filtros do site
// @Description Atualiza as palavras-chave de filtragem para um site inscrito
// @Tags UserSite
// @Accept json
// @Produce json
// @Param siteId path string true "ID do site"
// @Param body body model.UpdateUserSiteFiltersRequest true "Palavras-chave"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /userSite/{siteId} [patch]
```

#### 4.5 — plan_controller.go

**GetAllPlans:**
```go
// GetAllPlans godoc
// @Summary Listar planos
// @Description Retorna todos os planos de assinatura disponiveis
// @Tags Plans
// @Produce json
// @Success 200 {array} model.Plan
// @Failure 500 {object} model.ErrorResponse
// @Router /api/plans [get]
```

#### 4.6 — site_career_controller.go

**GetAllSites:**
```go
// GetAllSites godoc
// @Summary Listar sites de carreiras
// @Description Retorna todos os sites com status de inscricao do usuario
// @Tags Sites
// @Produce json
// @Success 200 {array} model.SiteDTO
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/getSites [get]
```

**InsertNewSiteCareer:**
```go
// InsertNewSiteCareer godoc
// @Summary Adicionar site de carreiras
// @Description Cria nova configuracao de scraping para um site (admin)
// @Tags Sites
// @Accept multipart/form-data
// @Produce json
// @Param siteData formData string true "JSON da configuracao do site"
// @Param logo formData file false "Logo da empresa"
// @Success 201 {object} model.SiteScrapingConfig
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /siteCareer [post]
```

**SandboxScrape:**
```go
// SandboxScrape godoc
// @Summary Testar scraping (sandbox)
// @Description Executa scraping de teste com a configuracao fornecida (admin)
// @Tags Sites
// @Accept json
// @Produce json
// @Param body body model.SiteScrapingConfig true "Configuracao de scraping"
// @Success 200 {object} model.SandboxScrapeResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.SandboxScrapeErrorResponse
// @Security CookieAuth
// @Router /scrape-sandbox [post]
```

#### 4.7 — payment_controller.go

**CreatePayment:**
```go
// CreatePayment godoc
// @Summary Criar pagamento
// @Description Inicia processo de pagamento para um plano
// @Tags Payment
// @Accept json
// @Produce json
// @Param planId path int true "ID do plano"
// @Param body body gateway.InitiatePaymentRequest true "Dados de pagamento"
// @Success 200 {object} model.CreatePaymentResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Failure 502 {object} model.ErrorResponse
// @Router /api/payments/create/{planId} [post]
```

**HandleWebhook:**
```go
// HandleWebhook godoc
// @Summary Webhook de pagamento
// @Description Recebe notificacao de pagamento do AbacatePay
// @Tags Payment
// @Accept json
// @Produce json
// @Param body body gateway.WebhookPayload true "Payload do webhook"
// @Success 200 {object} model.StatusResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/webhooks/abacatepay [post]
```

#### 4.8 — dashboardDataController.go

**GetDashboardDataByUserId:**
```go
// GetDashboardDataByUserId godoc
// @Summary Dados do dashboard
// @Description Retorna estatisticas e vagas recentes do usuario
// @Tags Dashboard
// @Produce json
// @Success 200 {object} model.DashboardData
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/dashboard [get]
```

**GetLatestJobs:**
```go
// GetLatestJobs godoc
// @Summary Vagas recentes paginadas
// @Description Retorna vagas com paginacao e filtros
// @Tags Dashboard
// @Produce json
// @Param page query int false "Pagina" default(1)
// @Param limit query int false "Limite por pagina (max 50)" default(10)
// @Param days query int false "Filtrar por dias" default(0)
// @Param search query string false "Buscar por titulo"
// @Success 200 {object} model.PaginatedJobs
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/dashboard/jobs [get]
```

#### 4.9 — health_controller.go

**Liveness:**
```go
// Liveness godoc
// @Summary Liveness check
// @Description Verifica se a API esta rodando
// @Tags Health
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Router /health/live [get]
```

**Readiness:**
```go
// Readiness godoc
// @Summary Readiness check
// @Description Verifica se o banco de dados e Redis estao acessiveis
// @Tags Health
// @Produce json
// @Success 200 {object} model.ReadinessResponse
// @Failure 503 {object} model.ReadinessResponse
// @Router /health/ready [get]
```

#### 4.10 — notification_controller.go

**GetNotificationsByUser:**
```go
// GetNotificationsByUser godoc
// @Summary Listar notificacoes
// @Description Retorna historico de notificacoes do usuario
// @Tags Notifications
// @Produce json
// @Param limit query int false "Limite de resultados (max 200)" default(50)
// @Success 200 {array} model.NotificationWithJob
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/notifications [get]
```

#### 4.11 — analysis_controller.go

**AnalyzeJob:**
```go
// AnalyzeJob godoc
// @Summary Analisar vaga com IA
// @Description Analisa compatibilidade do curriculo com uma vaga usando IA
// @Tags Analysis
// @Accept json
// @Produce json
// @Param body body model.AnalyzeJobRequest true "ID da vaga"
// @Success 200 {object} model.ResumeAnalysis
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/analyze-job [post]
```

**SendAnalysisEmail:**
```go
// SendAnalysisEmail godoc
// @Summary Enviar analise por email
// @Description Envia resultado da analise de compatibilidade por email
// @Tags Analysis
// @Accept json
// @Produce json
// @Param body body model.SendAnalysisEmailRequest true "ID da vaga e resultado da analise"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/analyze-job/send-email [post]
```

#### 4.12 — requested_site_controller.go

**Create:**
```go
// Create godoc
// @Summary Solicitar novo site
// @Description Envia solicitacao de um novo site de carreiras para o admin
// @Tags RequestedSite
// @Accept json
// @Produce json
// @Param body body model.RequestedSiteRequest true "URL do site"
// @Success 201 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/request-site [post]
```

#### 4.13 — admin_dashboard_controller.go

**GetAdminDashboard:**
```go
// GetAdminDashboard godoc
// @Summary Dashboard administrativo
// @Description Retorna metricas administrativas (somente admin)
// @Tags Admin
// @Produce json
// @Success 200 {object} model.AdminDashboardData
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/admin/dashboard [get]
```

**Step 2: Verify compilation**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go build ./...
```

Expected: may fail due to missing `docs/swagger` — that's expected, fixed in Task 5.

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add controller/
git commit -m "feat: add Swagger annotations to all 13 controllers (25 endpoints)"
```

---

### Task 5: Generate Swagger spec and verify

**Files:**
- Create (generated): `docs/swagger/docs.go`
- Create (generated): `docs/swagger/swagger.json`
- Create (generated): `docs/swagger/swagger.yaml`

**Step 1: Run swag init**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
swag init -g cmd/api/main.go -o docs/swagger
```

Expected: generates 3 files in `docs/swagger/`

**Step 2: Verify full project compiles**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go build ./...
```

Expected: SUCCESS

**Step 3: Run tests to make sure nothing is broken**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go test ./... -count=1
```

Expected: all tests pass

**Step 4: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add docs/swagger/ cmd/api/main.go
git commit -m "feat: generate Swagger spec and serve UI at /swagger (dev only)"
```

---

### Task 6: Manual verification

**Step 1: Start the API (requires database and Redis running)**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
go run ./cmd/api/main.go
```

**Step 2: Open Swagger UI**

Navigate to `http://localhost:8080/swagger/index.html`

**Step 3: Verify all endpoints are listed**

Check these tags appear with their endpoints:
- Auth (4): SignIn, SignUp, Logout, ValidateCheckout
- User (2): UpdateProfile, ChangePassword
- Curriculum (4): Create, Get, Update, SetActive
- Dashboard (2): GetDashboardData, GetLatestJobs
- Sites (3): GetAllSites, InsertNewSiteCareer, SandboxScrape
- UserSite (3): Insert, Delete, UpdateFilters
- Plans (1): GetAllPlans
- Payment (2): CreatePayment, HandleWebhook
- Analysis (2): AnalyzeJob, SendAnalysisEmail
- Notifications (1): GetNotificationsByUser
- RequestedSite (1): Create
- Admin (1): GetAdminDashboard
- Health (2): Liveness, Readiness

**Step 4: Verify Swagger is NOT served in release mode**

```bash
GIN_MODE=release go run ./cmd/api/main.go
# curl http://localhost:8080/swagger/index.html should return 404
```
