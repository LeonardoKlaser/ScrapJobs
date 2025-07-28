package main

import (
	"context"
	"log"
	"os"
	"web-scrapper/infra/db"
	"web-scrapper/infra/gemini"
	"web-scrapper/infra/ses"
	"web-scrapper/processor"
	"web-scrapper/repository"
	"web-scrapper/middleware"
	"web-scrapper/tasks"
	"web-scrapper/usecase"
	"web-scrapper/utils"
	"web-scrapper/model"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)



func main() {
	if os.Getenv("GIN_MODE") != "release"{
		godotenv.Load()	
	}

	var err error
	var secrets *model.AppSecrets

	secretName := os.Getenv("APP_SECRET_NAME")
	if secretName  != ""{
		secrets, err = utils.GetAppSecrets(secretName)
		if err != nil {
            panic("Failed to get secrets from AWS Secrets Manager: " + err.Error())
        }
	} else {
        secrets = &model.AppSecrets{
            DBHost:     os.Getenv("HOST_DB"),
            DBPort:     os.Getenv("PORT_DB"),
            DBUser: os.Getenv("USER_DB"),
            DBPassword: os.Getenv("PASSWORD_DB"),
            DBName:   os.Getenv("DBNAME"),
            RedisAddr: os.Getenv("REDIS_ADDR"),
			GeminiKey: os.Getenv("GEMINI_KEY"),
			AIModel: os.Getenv("AI_MODEL"),
        }
    }
	
	dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort,secrets.DBUser,secrets.DBPassword,secrets.DBName)
	if err != nil {
		middleware.Logger.Fatal().Err(err).Msg("Could not connect to database")
	}

	geminiConfig := gemini.Config{
		ApiKey:   secrets.GeminiKey,
		ApiModel: secrets.AIModel,
	}
	geminiClient, err := gemini.GeminiClientModel(context.Background(), geminiConfig)
	if err != nil {
		middleware.Logger.Fatal().Err(err).Msg("could not create gemini client")
	}
	
	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		middleware.Logger.Fatal().Err(err).Msg("could not load aws config")
	}
	clientSES := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(clientSES, "leobkklaser@gmail.com")

	
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
	NotificationRepository := repository.NewNotificationRepository(dbConnection)
	
	// Services & Usecases
	aiAnalyser := usecase.NewAiAnalyser(geminiClient)
	emailService := usecase.NewSESSenderAdapter(mailSender)
	jobUsecase := usecase.NewJobUseCase(jobRepository)

	aiApiLimiter := rate.NewLimiter(rate.Limit(30.0/60.0), 1)

	rateLimitedAiService := usecase.NewRateLimitedAiAnalyser(aiAnalyser, aiApiLimiter)

	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, rateLimitedAiService, emailService, NotificationRepository)
	
	// TaskProcessor 
	taskProcessor := processor.NewTaskProcessor(*jobUsecase, *notificationUsecase, clientAsynq, mailSender)
	

	// --- Mapeamento das Tarefas para os Handlers ---
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeScrapSite, taskProcessor.HandleScrapeSiteTask)
	mux.HandleFunc(tasks.TypeProcessResults, taskProcessor.HandleProcessResultsTask)

	log.Println("Worker Server started...")
	if err := srv.Run(middleware.AsynqMiddleware(mux)); err != nil {
		middleware.Logger.Fatal().Err(err).Msg("Could not run asynq server")
	}
}