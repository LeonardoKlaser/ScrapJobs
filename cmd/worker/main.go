package main

import (
	"context"
	"os"
	"web-scrapper/gateway"
	"web-scrapper/infra/db"
	"web-scrapper/infra/gemini"
	"web-scrapper/infra/ses"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/processor"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
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
			GeminiKey:  os.Getenv("GEMINI_KEY"),
			AIModel:    os.Getenv("AI_MODEL"),
		}
	}

	dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort, secrets.DBUser, secrets.DBPassword, secrets.DBName)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not connect to database")
	}
	defer dbConnection.Close()

	geminiConfig := gemini.Config{
		ApiKey:   secrets.GeminiKey,
		ApiModel: secrets.AIModel,
	}
	geminiClient, err := gemini.GeminiClientModel(context.Background(), geminiConfig)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("could not create gemini client")
	}

	// Carrega configuração AWS para SES (email)
	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("could not load aws config — email via SES não estará disponível")
	}

	senderEmail := os.Getenv("SES_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@scrapjobs.com.br"
	}

	clientSES := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(clientSES, senderEmail)
	emailService := usecase.NewSESSenderAdapter(mailSender)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: secrets.RedisAddr},
		asynq.Config{
			Concurrency: 4,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: secrets.RedisAddr})
	defer clientAsynq.Close()

	// Repositories
	jobRepository := repository.NewJobRepository(dbConnection)
	userSiteRepository := repository.NewUserSiteRepository(dbConnection)
	notificationRepository := repository.NewNotificationRepository(dbConnection)
	userRepository := repository.NewUserRepository(dbConnection)
	planRepository := repository.NewPlanRepository(dbConnection)
	dashboardRepository := repository.NewDashboardRepository(dbConnection)

	// Services & Usecases
	aiAnalyser := usecase.NewAiAnalyser(geminiClient)
	jobUsecase := usecase.NewJobUseCase(jobRepository)

	aiApiLimiter := rate.NewLimiter(rate.Limit(15.0/60.0), 1)
	rateLimitedAiService := usecase.NewRateLimitedAiAnalyser(aiAnalyser, aiApiLimiter)

	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, rateLimitedAiService, emailService, notificationRepository, clientAsynq, planRepository)

	// PaymentUsecase (necessário para HandleCompleteRegistrationTask)
	abacatepayGateway := gateway.NewAbacatePayGateway()
	userUsecase := usecase.NewUserUsercase(userRepository)
	redisOpt := asynq.RedisClientOpt{Addr: secrets.RedisAddr}
	paymentUsecase := usecase.NewPaymentUsecase(abacatepayGateway, redisOpt, userUsecase, planRepository)

	// TaskProcessor
	taskProcessor := processor.NewTaskProcessor(
		*jobUsecase,
		*notificationUsecase,
		paymentUsecase,
		emailService,
		clientAsynq,
		dashboardRepository,
	)

	// Mapeamento das Tarefas para os Handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeScrapSite, taskProcessor.HandleScrapeSiteTask)
	mux.HandleFunc(tasks.TypeProcessResults, taskProcessor.HandleFindMatchesTask)
	mux.HandleFunc(tasks.TypeAnalyzeUserJob, taskProcessor.HandleAnalyzeJobUserTask)
	mux.HandleFunc(tasks.TypeNotifyUser, taskProcessor.HandleNotifyTask)
	mux.HandleFunc(tasks.TypeCompleteRegistration, taskProcessor.HandleCompleteRegistrationTask)

	logging.Logger.Info().Msg("Worker Server started...")
	if err := srv.Run(logging.AsynqMiddleware(mux)); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not run asynq server")
	}
}
