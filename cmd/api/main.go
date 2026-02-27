package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-scrapper/controller"
	"web-scrapper/gateway"
	"web-scrapper/infra/db"
	"web-scrapper/infra/openai"
	"web-scrapper/infra/metrics"
	redispkg "web-scrapper/infra/redis"
	"web-scrapper/infra/s3"
	"web-scrapper/infra/ses"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/middleware"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"

	_ "web-scrapper/docs/swagger"
)

// @title ScrapJobs API
// @version 1.0
// @description API para plataforma de scraping e matching de vagas de emprego.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name Authorization
func main() {
	if os.Getenv("GIN_MODE") != "release" {
		godotenv.Load()
	}

	jwtSecret := os.Getenv("JWTTOKEN")
	if len(jwtSecret) < 32 {
		logging.Logger.Fatal().Msg("JWTTOKEN environment variable must be at least 32 characters")
	}

	server := gin.Default()

	allowedOrigins := []string{"http://localhost:5173", "https://scrapjobs.com.br"}
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		isDuplicate := false
		for _, o := range allowedOrigins {
			if o == frontendURL {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			allowedOrigins = append(allowedOrigins, frontendURL)
		}
	}

	server.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	var err error
	var secrets *model.AppSecrets

	secretName := os.Getenv("APP_SECRET_NAME")
	if secretName != "" {
		secrets, err = utils.GetAppSecrets(secretName)
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("Could not get secrets from AWS Secrets Manager")
		}
	} else {
		secrets = &model.AppSecrets{
			DBHost:     os.Getenv("HOST_DB"),
			DBPort:     os.Getenv("PORT_DB"),
			DBUser:     os.Getenv("USER_DB"),
			DBPassword: os.Getenv("PASSWORD_DB"),
			DBName:     os.Getenv("DBNAME"),
			RedisAddr:  os.Getenv("REDIS_ADDR"),
			OpenAIKey:  os.Getenv("OPENAI_API_KEY"),
			AIModel:    os.Getenv("AI_MODEL"),
		}
	}

	if err := utils.ValidateSecrets(secrets); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Invalid configuration")
	}

	var dbConnection *sql.DB
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dbConnection, err = db.ConnectDBFromURL(dbURL)
	} else {
		dbConnection, err = db.ConnectDB(secrets.DBHost, secrets.DBPort, secrets.DBUser, secrets.DBPassword, secrets.DBName)
	}
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not connect to database")
	}
	logging.Logger.Info().Msg("successfully connected to the database")
	defer dbConnection.Close()

	redisAddr := secrets.RedisAddr
	if redisAddr == "" {
		redisAddr = os.Getenv("REDIS_URL")
	}

	redisClient, err := redispkg.NewRedisClient(redisAddr)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not connect to Redis")
	}
	defer redisClient.Close()
	logging.Logger.Info().Msg("Connected to Redis")

	var asynqRedisOpt asynq.RedisConnOpt = asynq.RedisClientOpt{Addr: redisAddr}
	if parsed, parseErr := asynq.ParseRedisURI(redisAddr); parseErr == nil {
		asynqRedisOpt = parsed
	}
	asynqClient := asynq.NewClient(asynqRedisOpt)
	defer asynqClient.Close()

	// --- Prometheus pool collectors ---
	metrics.RegisterDBCollector(dbConnection)
	metrics.RegisterRedisCollector(redisClient)

	// --- S3 Uploader (opcional — usado apenas para upload de logos de sites) ---
	var s3Uploader s3.UploaderInterface
	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	if s3BucketName != "" {
		awsCfg, awsErr := awsconfig.LoadDefaultConfig(context.TODO())
		if awsErr != nil {
			logging.Logger.Warn().Err(awsErr).Msg("Falha ao carregar configuração AWS — upload de logos via S3 desabilitado")
		} else {
			s3Uploader = s3.NewUploader(awsCfg, s3BucketName)
			logging.Logger.Info().Str("bucket", s3BucketName).Msg("S3 uploader configurado")
		}
	} else {
		logging.Logger.Warn().Msg("S3_BUCKET_NAME não definida — upload de logos via S3 desabilitado")
		s3Uploader = &s3.NoOpUploader{}
	}

	// --- SES Email Sender ---
	senderEmail := os.Getenv("SES_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@scrapjobs.com.br"
	}

	awsCfgSES, sesErr := ses.LoadAWSConfig(context.Background())
	if sesErr != nil {
		logging.Logger.Warn().Err(sesErr).Msg("Falha ao carregar configuração AWS para SES — envio de email pode não funcionar")
	}
	clientSES := ses.LoadAWSClient(awsCfgSES)
	mailSender := ses.NewSESMailSender(clientSES, senderEmail)
	emailService := usecase.NewSESSenderAdapter(mailSender)

	// --- Gateway de Pagamento ---
	abacatepayGateway := gateway.NewAbacatePayGateway()

	// --- OpenAI Client (opcional — usado para análise manual de vagas) ---
	var aiAnalyser interfaces.AnalysisService
	if secrets.OpenAIKey != "" && secrets.AIModel != "" {
		openaiConfig := openai.Config{
			ApiKey:   secrets.OpenAIKey,
			ApiModel: secrets.AIModel,
		}
		openaiClient, openaiErr := openai.NewOpenAIClient(openaiConfig)
		if openaiErr != nil {
			logging.Logger.Warn().Err(openaiErr).Msg("Falha ao criar cliente OpenAI — análise manual de IA desabilitada")
		} else {
			aiAnalyser = usecase.NewAiAnalyser(openaiClient)
			logging.Logger.Info().Msg("Cliente OpenAI configurado para análise manual de IA")
		}
	} else {
		logging.Logger.Warn().Msg("OPENAI_API_KEY ou AI_MODEL não definidos — análise manual de IA desabilitada")
	}

	// Repositories
	userRepository := repository.NewUserRepository(dbConnection)
	curriculumRepository := repository.NewCurriculumRepository(dbConnection)
	userSiteRepository := repository.NewUserSiteRepository(dbConnection)
	siteCareerRepository := repository.NewSiteCareerRepository(dbConnection)
	dashboardRepository := repository.NewDashboardRepository(dbConnection)
	planRepository := repository.NewPlanRepository(dbConnection)
	requestedSiteRepository := repository.NewRequestedSiteRepository(dbConnection)
	notificationRepository := repository.NewNotificationRepository(dbConnection)
	jobRepository := repository.NewJobRepository(dbConnection)
	passwordResetRepo := repository.NewPasswordResetRepository(dbConnection)

	// Usecases
	userUsecase := usecase.NewUserUsercase(userRepository)
	curriculumUsecase := usecase.NewCurriculumUsecase(curriculumRepository)
	userSiteUsecase := usecase.NewUserSiteUsecase(userSiteRepository, planRepository)
	siteCareerUsecase := usecase.NewSiteCareerUsecase(siteCareerRepository, s3Uploader)
	planUsecase := usecase.NewPlanUsecase(planRepository)
	requestedSiteUsecase := usecase.NewRequestedSiteUsecase(requestedSiteRepository)
	paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, redisClient, userUsecase, planRepository)
	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, nil, emailService, notificationRepository, planRepository, userRepository)

	// Controllers
	userController := controller.NewUserController(userUsecase)
	curriculumController := controller.NewCurriculumController(curriculumUsecase)
	userSiteController := controller.NewUserSiteController(userSiteUsecase)
	siteCareerController := controller.NewSiteCareerController(siteCareerUsecase, userSiteRepository)
	healthController := controller.NewHealthController(dbConnection, asynqClient, redisClient)
	checkAuthController := controller.NewCheckAuthController(userSiteRepository)
	dashboardController := controller.NewDashboardDataController(dashboardRepository)
	planController := controller.NewPlanController(planUsecase)
	requestedSiteController := controller.NewRequestedSiteController(requestedSiteUsecase)
	paymentController := controller.NewPaymentController(paymentUsecase, emailService, asynqClient)
	notificationController := controller.NewNotificationController(notificationUsecase)

	passwordResetController := controller.NewPasswordResetController(passwordResetRepo, userRepository, emailService)

	accountController := controller.NewAccountController(userRepository)

	adminDashboardController := controller.NewAdminDashboardController(dashboardRepository)

	// Analysis Controller (análise manual de IA)
	var analysisController *controller.AnalysisController
	if aiAnalyser != nil {
		analysisController = controller.NewAnalysisController(aiAnalyser, curriculumRepository, jobRepository, notificationRepository, planRepository, emailService)
	}

	// Middleware
	middlewareAuth := middleware.NewMiddleware(userUsecase)

	// Rate limiters — distributed via Redis
	rateLimiterFn := newRedisRateLimiterFactory(redisClient)
	publicRateLimiter := rateLimiterFn(5, 60)

	csrfMiddleware := middleware.CSRFProtection()

	publicRoutes := server.Group("/")
	publicRoutes.Use(logging.GinMiddleware())
	publicRoutes.Use(metrics.GinPrometheus())
	publicRoutes.Use(csrfMiddleware)
	publicRoutes.Use(publicRateLimiter)
	{
		publicRoutes.POST("/login", userController.SignIn)
		publicRoutes.GET("/api/plans", planController.GetAllPlans)
		publicRoutes.POST("/api/payments/create/:planId", paymentController.CreatePayment)
		publicRoutes.POST("/api/webhooks/abacatepay", utils.WebhookAuthMiddleware(), paymentController.HandleWebhook)
	}

	// Forgot/reset password — rate limiter próprio, mais restritivo
	forgotPasswordLimiter := rateLimiterFn(3, 60)
	forgotPasswordRoutes := server.Group("/")
	forgotPasswordRoutes.Use(logging.GinMiddleware())
	forgotPasswordRoutes.Use(metrics.GinPrometheus())
	forgotPasswordRoutes.Use(csrfMiddleware)
	forgotPasswordRoutes.Use(forgotPasswordLimiter)
	{
		forgotPasswordRoutes.POST("/api/forgot-password", passwordResetController.ForgotPassword)
		forgotPasswordRoutes.POST("/api/reset-password", passwordResetController.ResetPassword)
	}

	// Checkout validation — rate limiter próprio, separado do publicRoutes para não herdar o 5/min
	checkoutValidationLimiter := rateLimiterFn(10, 60)
	checkoutRoutes := server.Group("/")
	checkoutRoutes.Use(logging.GinMiddleware())
	checkoutRoutes.Use(metrics.GinPrometheus())
	checkoutRoutes.Use(csrfMiddleware)
	checkoutRoutes.Use(checkoutValidationLimiter)
	{
		checkoutRoutes.POST("/api/users/validate-checkout", userController.ValidateCheckout)
	}

	privateRateLimiter := rateLimiterFn(15, 60)

	// Routes accessible even when subscription is expired
	privateRoutes := server.Group("/")
	privateRoutes.Use(logging.GinMiddleware(), metrics.GinPrometheus(), csrfMiddleware, middlewareAuth.RequireAuth)
	{
		privateRoutes.GET("api/me", checkAuthController.CheckAuthUser)
		privateRoutes.POST("/api/logout", userController.Logout)
		privateRoutes.PATCH("/api/user/profile", userController.UpdateProfile)
		privateRoutes.POST("/api/user/change-password", userController.ChangePassword)
		privateRoutes.DELETE("/api/user/account", accountController.DeleteAccount)
	}

	// Routes that require active subscription
	subscribedRoutes := server.Group("/")
	subscribedRoutes.Use(logging.GinMiddleware(), metrics.GinPrometheus(), csrfMiddleware, middlewareAuth.RequireAuth, middleware.RequireActiveSubscription(), privateRateLimiter)
	{
		subscribedRoutes.GET("api/dashboard", dashboardController.GetDashboardDataByUserId)
		subscribedRoutes.GET("api/dashboard/jobs", dashboardController.GetLatestJobs)
		subscribedRoutes.GET("api/getSites", siteCareerController.GetAllSites)
		subscribedRoutes.GET("api/notifications", notificationController.GetNotificationsByUser)
		subscribedRoutes.POST("/curriculum", curriculumController.CreateCurriculum)
		subscribedRoutes.PUT("/curriculum/:id", curriculumController.UpdateCurriculum)
		subscribedRoutes.DELETE("/curriculum/:id", curriculumController.DeleteCurriculum)
		subscribedRoutes.GET("/curriculum", curriculumController.GetCurriculumByUserId)
		subscribedRoutes.POST("/userSite", userSiteController.InsertUserSite)
		subscribedRoutes.DELETE("/userSite/:siteId", userSiteController.DeleteUserSite)
		subscribedRoutes.PATCH("/userSite/:siteId", userSiteController.UpdateUserSiteFilters)
		subscribedRoutes.POST("api/request-site", requestedSiteController.Create)
		if analysisController != nil {
			analyzeRateLimiter := rateLimiterFn(3, 60)
			subscribedRoutes.POST("/api/analyze-job", analyzeRateLimiter, analysisController.AnalyzeJob)
			subscribedRoutes.POST("/api/analyze-job/send-email", analyzeRateLimiter, analysisController.SendAnalysisEmail)
			subscribedRoutes.GET("/api/analyze-job/history", analysisController.GetAnalysisHistory)
		}
	}

	// Admin routes — require authentication + admin role
	adminRoutes := server.Group("/")
	adminRoutes.Use(logging.GinMiddleware())
	adminRoutes.Use(metrics.GinPrometheus())
	adminRoutes.Use(csrfMiddleware)
	adminRoutes.Use(middlewareAuth.RequireAuth)
	adminRoutes.Use(middleware.RequireAdmin())
	{
		adminRoutes.GET("/api/admin/dashboard", adminDashboardController.GetAdminDashboard)
		adminRoutes.POST("/siteCareer", siteCareerController.InsertNewSiteCareer)
		adminRoutes.POST("/scrape-sandbox", siteCareerController.SandboxScrape)
	}

	healthRoutes := server.Group("/health")
	{
		healthRoutes.GET("/live", healthController.Liveness)
		healthRoutes.GET("/ready", healthController.Readiness)
	}

	// Swagger documentation (dev only)
	if os.Getenv("GIN_MODE") != "release" {
		server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Prometheus metrics endpoint
	server.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Logger.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	logging.Logger.Info().Msg("Server exited gracefully")
}

// newRedisRateLimiterFactory returns a function that creates rate-limiting
// middleware backed by Redis. If redisClient is nil, it falls back to in-memory.
func newRedisRateLimiterFactory(redisClient *redis.Client) func(limit, windowSeconds int) gin.HandlerFunc {
	if redisClient == nil {
		return func(limit, windowSeconds int) gin.HandlerFunc {
			// Convert to rate.Limit: limit requests per windowSeconds
			r := rate.Limit(float64(limit) / float64(windowSeconds))
			return middleware.RateLimiter(r, limit)
		}
	}
	return func(limit, windowSeconds int) gin.HandlerFunc {
		return middleware.RedisRateLimiter(redisClient, limit, windowSeconds)
	}
}
