package main

import (
	"context"
	"log"
	"os"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/infra/gemini"
	"web-scrapper/infra/ses"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"web-scrapper/usecase"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

// just to remember to change to env variable
const redisAddr = "redis:6379"

func main() {
	godotenv.Load()
	dbConnection, err := db.ConnectDB(os.Getenv("HOST_DB"), os.Getenv("PORT_DB"), os.Getenv("USER_DB"), os.Getenv("PASSWORD_DB"), os.Getenv("DBNAME"))
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}

	geminiConfig := gemini.Config{
		ApiKey:   os.Getenv("GEMINI_KEY"),
		ApiModel: os.Getenv("AI_MODEL"),
	}
	geminiClient, err := gemini.GeminiClientModel(context.Background(), geminiConfig)
	if err != nil {
		log.Fatalf("could not create gemini client: %v", err)
	}
	
	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		log.Fatalf("could not load aws config: %v", err)
	}
	clientSES := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(clientSES, "leobkklaser@gmail.com")

	// O Worker precisa do SERVIDOR Asynq
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
    defer clientAsynq.Close()
	// --- Injeção de Dependência para os Handlers ---
	
	// Repositories
	jobRepository := repository.NewJobRepository(dbConnection)
	userSiteRepository := repository.NewUserSiteRepository(dbConnection)
	curriculumRepository := repository.NewCurriculumRepository(dbConnection)
	NotificationRepository := repository.NewNotificationRepository(dbConnection)
	
	// Services & Usecases
	aiAnalyser := usecase.NewAiAnalyser(geminiClient)
	emailService := usecase.NewSESSenderAdapter(mailSender)
	jobUsecase := usecase.NewJobUseCase(jobRepository)
	notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, curriculumRepository, aiAnalyser, emailService, *NotificationRepository)
	
	// O TaskProcessor é o coração do nosso worker
	taskProcessor := controller.NewTaskProcessor(*jobUsecase, *notificationUsecase, clientAsynq)


	// --- Mapeamento das Tarefas para os Handlers ---
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeScrapSite, taskProcessor.HandleScrapeSiteTask)
	mux.HandleFunc(tasks.TypeProcessResults, taskProcessor.HandleProcessResultsTask)

	log.Println("Worker Server started...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run asynq server: %v", err)
	}
}