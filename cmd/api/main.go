package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-scrapper/controller"
	"web-scrapper/gateway"
	"web-scrapper/infra/db"
	"web-scrapper/infra/gemini"
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
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

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
			GeminiKey:  os.Getenv("GEMINI_KEY"),
			AIModel:    os.Getenv("AI_MODEL"),
		}
	}

	dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort, secrets.DBUser, secrets.DBPassword, secrets.DBName)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not connect to database")
	}
	logging.Logger.Info().Msg("successfully connected to the database")
	defer dbConnection.Close()

	redisOpt := connectRedis(secrets)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: secrets.RedisAddr})
	defer asynqClient.Close()

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

	// --- Gemini AI Client (opcional — usado para análise manual de vagas) ---
	var aiAnalyser interfaces.AnalysisService
	if secrets.GeminiKey != "" && secrets.AIModel != "" {
		geminiConfig := gemini.Config{
			ApiKey:   secrets.GeminiKey,
			ApiModel: secrets.AIModel,
		}
		geminiClient, geminiErr := gemini.GeminiClientModel(context.Background(), geminiConfig)
		if geminiErr != nil {
			logging.Logger.Warn().Err(geminiErr).Msg("Falha ao criar cliente Gemini — análise manual de IA desabilitada")
		} else {
			aiAnalyser = usecase.NewAiAnalyser(geminiClient)
			logging.Logger.Info().Msg("Cliente Gemini configurado para análise manual de IA")
		}
	} else {
		logging.Logger.Warn().Msg("GEMINI_KEY ou AI_MODEL não definidos — análise manual de IA desabilitada")
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

	// Usecases
	userUsecase := usecase.NewUserUsercase(userRepository)
	curriculumUsecase := usecase.NewCurriculumUsecase(curriculumRepository)
	userSiteUsecase := usecase.NewUserSiteUsecase(userSiteRepository, planRepository)
	siteCareerUsecase := usecase.NewSiteCareerUsecase(siteCareerRepository, s3Uploader)
	planUsecase := usecase.NewPlanUsecase(planRepository)
	requestedSiteUsecase := usecase.NewRequestedSiteUsecase(requestedSiteRepository)
	paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, redisOpt, userUsecase, planRepository)
	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, nil, emailService, notificationRepository, asynqClient, planRepository)

	// Controllers
	userController := controller.NewUserController(userUsecase)
	curriculumController := controller.NewCurriculumController(curriculumUsecase)
	userSiteController := controller.NewUserSiteController(userSiteUsecase)
	siteCareerController := controller.NewSiteCareerController(siteCareerUsecase, userSiteRepository)
	healthController := controller.NewHealthController(dbConnection, asynqClient)
	checkAuthController := controller.NewCheckAuthController()
	dashboardController := controller.NewDashboardDataController(dashboardRepository)
	planController := controller.NewPlanController(planUsecase)
	requestedSiteController := controller.NewRequestedSiteController(requestedSiteUsecase)
	paymentController := controller.NewPaymentController(paymentUsecase, emailService, asynqClient)
	notificationController := controller.NewNotificationController(notificationUsecase)

	adminDashboardController := controller.NewAdminDashboardController(dashboardRepository)

	// Analysis Controller (análise manual de IA)
	var analysisController *controller.AnalysisController
	if aiAnalyser != nil {
		analysisController = controller.NewAnalysisController(aiAnalyser, curriculumRepository, jobRepository, notificationRepository, planRepository)
	}

	// Middleware
	middlewareAuth := middleware.NewMiddleware(userUsecase)

	// Rate limiters
	publicRateLimiter := middleware.RateLimiter(rate.Limit(5.0/60.0), 2)

	publicRoutes := server.Group("/")
	publicRoutes.Use(logging.GinMiddleware())
	publicRoutes.Use(publicRateLimiter)
	{
		publicRoutes.POST("/login", userController.SignIn)
		publicRoutes.GET("/api/plans", planController.GetAllPlans)
		publicRoutes.POST("/api/payments/create/:planId", paymentController.CreatePayment)
		publicRoutes.POST("/api/webhooks/abacatepay", utils.WebhookAuthMiddleware(), paymentController.HandleWebhook)
	}

	privateRateLimiter := middleware.RateLimiter(rate.Limit(15.0/60.0), 10)

	privateRoutes := server.Group("/")
	privateRoutes.Use(logging.GinMiddleware())
	privateRoutes.Use(middlewareAuth.RequireAuth)
	{
		privateRoutes.GET("api/me", checkAuthController.CheckAuthUser)
		privateRoutes.GET("api/dashboard", dashboardController.GetDashboardDataByUserId)
		privateRoutes.GET("api/getSites", siteCareerController.GetAllSites)
		privateRoutes.GET("api/notifications", notificationController.GetNotificationsByUser)
		privateRoutes.GET("/api/admin/dashboard", adminDashboardController.GetAdminDashboard)
	}
	privateRoutes.Use(privateRateLimiter)
	{
		privateRoutes.POST("/curriculum", curriculumController.CreateCurriculum)
		privateRoutes.PUT("/curriculum/:id", curriculumController.UpdateCurriculum)
		privateRoutes.PATCH("/curriculum/:id/active", curriculumController.SetActiveCurriculum)
		privateRoutes.POST("/userSite", userSiteController.InsertUserSite)
		privateRoutes.DELETE("/userSite/:siteId", userSiteController.DeleteUserSite)
		privateRoutes.PATCH("/userSite/:siteId", userSiteController.UpdateUserSiteFilters)
		privateRoutes.POST("/siteCareer", siteCareerController.InsertNewSiteCareer)
		privateRoutes.POST("/scrape-sandbox", siteCareerController.SandboxScrape)
		privateRoutes.GET("/curriculum", curriculumController.GetCurriculumByUserId)
		privateRoutes.POST("/api/logout", userController.Logout)
		privateRoutes.PATCH("/api/user/profile", userController.UpdateProfile)
		privateRoutes.POST("/api/user/change-password", userController.ChangePassword)
		privateRoutes.POST("api/request-site", requestedSiteController.Create)
		if analysisController != nil {
			analyzeRateLimiter := middleware.RateLimiter(rate.Limit(3.0/60.0), 2)
			privateRoutes.POST("/api/analyze-job", analyzeRateLimiter, analysisController.AnalyzeJob)
		}
	}

	healthRoutes := server.Group("/health")
	{
		healthRoutes.GET("/live", healthController.Liveness)
		healthRoutes.GET("/ready", healthController.Readiness)
	}

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

func connectRedis(secrets *model.AppSecrets) asynq.RedisConnOpt {
	redisOpt := asynq.RedisClientOpt{Addr: secrets.RedisAddr}
	client := redisOpt.MakeRedisClient().(redis.UniversalClient)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		client.Close()
		logging.Logger.Fatal().Err(err).Str("redis_addr", secrets.RedisAddr).Msg("Falha ao conectar ao Redis")
	}
	client.Close()
	logging.Logger.Info().Str("redis_addr", secrets.RedisAddr).Msg("Conectado ao Redis com sucesso")
	return redisOpt
}
