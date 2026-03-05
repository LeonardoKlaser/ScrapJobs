package main

import (
	"context"
	"database/sql"
	"os"
	"web-scrapper/gateway"
	"web-scrapper/infra/db"
	redispkg "web-scrapper/infra/redis"
	"web-scrapper/infra/resend"
	"web-scrapper/infra/ses"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/processor"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("GIN_MODE") != "release" {
		godotenv.Load()
	}

	var err error
	var secrets *model.AppSecrets

	secretName := os.Getenv("APP_SECRET_NAME")
	if secretName != "" {
		secrets, err = utils.GetAppSecrets(secretName)
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("Failed to get secrets from AWS Secrets Manager")
		}
	} else {
		secrets = &model.AppSecrets{
			DBHost:     os.Getenv("HOST_DB"),
			DBPort:     os.Getenv("PORT_DB"),
			DBUser:     os.Getenv("USER_DB"),
			DBPassword: os.Getenv("PASSWORD_DB"),
			DBName:     os.Getenv("DBNAME"),
			RedisAddr:  os.Getenv("REDIS_ADDR"),
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
	defer dbConnection.Close()

	// --- Email Providers ---
	senderEmail := os.Getenv("SES_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@scrapjobs.com.br"
		logging.Logger.Warn().Msg("SES_SENDER_EMAIL nao definida — usando fallback noreply@scrapjobs.com.br")
	}

	emailSenders := make(map[string]interfaces.MailSender)

	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("could not load aws config — SES indisponível")
	} else {
		clientSES := ses.LoadAWSClient(awsCfg)
		emailSenders["ses"] = ses.NewSESMailSender(clientSES, senderEmail)
		logging.Logger.Info().
			Str("sender_email", senderEmail).
			Str("aws_region", awsCfg.Region).
			Msg("SES configurado com sucesso")
	}

	resendKey := os.Getenv("RESEND_API_KEY")
	resendFrom := os.Getenv("RESEND_SENDER_EMAIL")
	if resendFrom == "" {
		resendFrom = senderEmail
	}
	if resendKey != "" {
		emailSenders["resend"] = resend.NewResendMailSender(resendKey, resendFrom)
		logging.Logger.Info().Msg("Resend email sender configurado")
	} else {
		logging.Logger.Warn().Msg("RESEND_API_KEY não definida — Resend indisponível")
	}

	emailConfigRepo := repository.NewEmailConfigRepo(dbConnection)
	orchestrator := usecase.NewEmailOrchestrator(emailSenders, emailConfigRepo)
	emailService := usecase.NewSESSenderAdapter(orchestrator)

	redisAddr := secrets.RedisAddr
	if redisAddr == "" {
		redisAddr = os.Getenv("REDIS_URL")
	}

	var asynqRedisOpt asynq.RedisConnOpt = asynq.RedisClientOpt{Addr: redisAddr}
	if parsed, parseErr := asynq.ParseRedisURI(redisAddr); parseErr == nil {
		asynqRedisOpt = parsed
	}

	srv := asynq.NewServer(
		asynqRedisOpt,
		asynq.Config{
			Concurrency: 4,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	clientAsynq := asynq.NewClient(asynqRedisOpt)
	defer clientAsynq.Close()

	// Repositories
	jobRepository := repository.NewJobRepository(dbConnection)
	userSiteRepository := repository.NewUserSiteRepository(dbConnection)
	notificationRepository := repository.NewNotificationRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)
	planRepository := repository.NewPlanRepository(dbConnection)
	dashboardRepository := repository.NewDashboardRepository(dbConnection)

	// Services & Usecases
	jobUsecase := usecase.NewJobUseCase(jobRepository)

	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, nil, emailService, notificationRepository, planRepository, userRepository)

	// PaymentUsecase (necessário para HandleCompleteRegistrationTask)
	abacatepayGateway := gateway.NewAbacatePayGateway()
	userUsecase := usecase.NewUserUsercase(userRepository)
	workerRedisClient, err := redispkg.NewRedisClient(redisAddr)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not connect to Redis for PaymentUsecase")
	}
	defer workerRedisClient.Close()
	paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, workerRedisClient, userUsecase, planRepository)

	// TaskProcessor
	taskProcessor := processor.NewTaskProcessor(
		*jobUsecase,
		*notificationUsecase,
		paymentUsecase,
		emailService,
		dashboardRepository,
	)

	// Mapeamento das Tarefas para os Handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeScrapSite, taskProcessor.HandleScrapeSiteTask)
	mux.HandleFunc(tasks.TypeMatchUser, taskProcessor.HandleMatchUserTask)
	mux.HandleFunc(tasks.TypeSendDigest, taskProcessor.HandleSendDigestTask)
	mux.HandleFunc(tasks.TypeCompleteRegistration, taskProcessor.HandleCompleteRegistrationTask)

	logging.Logger.Info().Msg("Worker Server started...")
	if err := srv.Run(logging.AsynqMiddleware(mux)); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not run asynq server")
	}
}
