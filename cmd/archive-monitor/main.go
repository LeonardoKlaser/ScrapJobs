package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-scrapper/infra/db"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
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

	pollingInterval := 5 * time.Minute
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Archive monitor worker started. Polling interval:", pollingInterval)

	for {
		select {
		case <-ticker.C:
			log.Println("Polling for archived tasks...")
			// Lógica de verificação e notificação entra aqui
			// processArchivedTasks(ctx, inspector, notifier, redisClient)

		case sig := <-shutdown:
			log.Printf("Received shutdown signal: %v. Shutting down gracefully...", sig)
			// Lógica de finalização
			return
		}
	}
}