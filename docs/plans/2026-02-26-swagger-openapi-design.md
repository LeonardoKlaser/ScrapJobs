# Design: Swagger/OpenAPI Documentation — ScrapJobs API

## Objetivo

Documentar todos os endpoints REST com Swagger (OpenAPI 2.0), gerando spec automaticamente dos comentarios Go usando swaggo/swag.

## Dependencias

- `github.com/swaggo/swag/v2` (CLI + parser)
- `github.com/swaggo/gin-swagger` (middleware Gin)
- `github.com/swaggo/files` (Swagger UI embeddido)

## Componentes

### 1. Anotacao geral (cmd/api/main.go)

```go
// @title ScrapJobs API
// @version 1.0
// @description API para plataforma de scraping e matching de vagas
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name Authorization
```

### 2. DTOs tipados (model/dto.go)

Structs para request/response que hoje usam `gin.H{}` ou `map[string]`:

- LoginRequest, SignUpRequest
- StatusResponse, ErrorResponse, MessageResponse
- ValidateCheckoutRequest, ValidateCheckoutResponse
- ChangePasswordRequest, UpdateProfileRequest
- AnalyzeJobRequest, SendAnalysisEmailRequest, AnalyzeJobResponse, SendAnalysisEmailResponse
- CreatePaymentResponse
- HealthResponse, ReadinessResponse
- DashboardJobsResponse (wrapper para PaginatedJobs)
- InsertUserSiteResponse, RequestedSiteRequest
- SandboxScrapeResponse
- AdminDashboardResponse (wrapper)
- GetSitesResponse (array wrapper)

### 3. Anotacoes por controller (13 arquivos, ~25 endpoints)

Tags: Auth, User, Curriculum, Dashboard, Sites, UserSite, Payment, Analysis, Notification, Admin, Health, RequestedSite

Cada handler recebe: @Summary, @Tags, @Accept, @Produce, @Param, @Success, @Failure, @Security, @Router

### 4. Rota Swagger (so em dev)

```go
if os.Getenv("GIN_MODE") != "release" {
    server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

### 5. Geracao

```bash
swag init -g cmd/api/main.go -o docs/swagger
```

Gera: `docs/swagger/docs.go`, `swagger.json`, `swagger.yaml`

## Decisoes

- **DTOs para todos os endpoints** — mesmo respostas simples como `{"status": "ok"}` terao struct tipada
- **Nao altera logica dos handlers** — DTOs sao para documentacao, `gin.H{}` continua funcionando
- **Models existentes reutilizados** — User, Curriculum, Plan, Job, etc. ja tem json tags corretos
- **Swagger so em dev** — protegido por `GIN_MODE != "release"`

## Endpoints a documentar

| Controller | Endpoints |
|---|---|
| UserController | SignIn, SignUp, Logout, UpdateProfile, ChangePassword, ValidateCheckout |
| CheckAuthController | CheckAuthUser |
| CurriculumController | Create, Update, SetActive, GetByUserId |
| DashboardController | GetDashboardData, GetLatestJobs |
| SiteCareerController | GetAllSites, InsertNewSiteCareer, SandboxScrape |
| UserSiteController | Insert, Delete, UpdateFilters |
| PlanController | GetAllPlans |
| PaymentController | CreatePayment, HandleWebhook |
| AnalysisController | AnalyzeJob, SendAnalysisEmail |
| NotificationController | GetNotificationsByUser |
| RequestedSiteController | Create |
| AdminDashboardController | GetAdminDashboard |
| HealthController | Liveness, Readiness |
