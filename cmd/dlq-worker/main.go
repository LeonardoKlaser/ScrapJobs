package main

import(
	"context"
	"log"
	"os"
	"web-scrapper/infra/db"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
	"web-scrapper/processor"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main(){
	if os.Getenv("GIN_MODE") != "release"{
		godotenv.Load()
	}

	var err error
	var secrets *model.AppSecrets

	secretName := os.Getenv("APP_SECRET_NAME")
	if secretName != "" {
		secrets, err = utils.GetAppSecrets(secretName)
		if err != nil {
			panic("Failed to get secrets from AWS Secrets Manager: " + err.Error())
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

	awsCfg, err := ses.LoadAWSConfig(context.Background())
	if err != nil {
		log.Fatalf("could not load aws config: %v", err)
	}
	clientSES := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(clientSES, "admin@scrapjobs.com.br")
	emailService := usecase.NewSESSenderAdapter(mailSender)

	dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort, secrets.DBUser, secrets.DBPassword, secrets.DBName)
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}
	clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: secrets.RedisAddr})
	defer clientAsynq.Close()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{ Addr: secrets.RedisAddr},
		asynq.Config{
			Concurrency: 1,
			Queues: map[string]int{
				"dead" : 1,
			},
		},
	)

	taskProcessor := processor.NewTaskProcessor(usecase.JobUseCase{}, usecase.NotificationsUsecase{}, clientAsynq, emailService)

	mux := asynq.NewServeMux()

	
	//mux.HandleFunc(, taskProcessor.HandleDeadQueueLetter())

	log.Println("DLQ Worker Server started...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run DLQ worker server: %v", err)
	}
}