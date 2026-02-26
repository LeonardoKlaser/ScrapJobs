package main

import (
	"context"
	"os"
	"web-scrapper/gateway"
	"web-scrapper/infra/db"
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

	dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort, secrets.DBUser, secrets.DBPassword, secrets.DBName)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not connect to database")
	}
	defer dbConnection.Close()

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
	jobUsecase := usecase.NewJobUseCase(jobRepository)

	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, nil, emailService, notificationRepository, clientAsynq, planRepository, userRepository)

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
	mux.HandleFunc(tasks.TypeMatchUser, taskProcessor.HandleMatchUserTask)
	mux.HandleFunc(tasks.TypeSendDigest, taskProcessor.HandleSendDigestTask)
	mux.HandleFunc(tasks.TypeCompleteRegistration, taskProcessor.HandleCompleteRegistrationTask)

	logging.Logger.Info().Msg("Worker Server started...")
	if err := srv.Run(logging.AsynqMiddleware(mux)); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Could not run asynq server")
	}
}
