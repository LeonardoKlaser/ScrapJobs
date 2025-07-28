package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"web-scrapper/infra/ses"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/redis/go-redis/v9"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Archive Monitor service...")
	cfg, err := utils.LoadMonitorConfig()
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize dependencies using the loaded config
	redisOpt, err := redis.ParseURL(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("FATAL: could not parse redis address: %v", err)
	}
	redisClient := redis.NewClient(redisOpt)
	defer redisClient.Close()

	asynqRedisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}
	inspector := asynq.NewInspector(asynqRedisOpt)
	defer inspector.Close()

	awsCfg, err := ses.LoadAWSConfig(ctx)
	if err != nil {
		log.Fatalf("FATAL: could not load aws config: %v", err)
	}
	sesClient := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(sesClient, cfg.SenderEmail)

	// Instantiate architectural components
	monitorRepo := repository.NewMonitorRepository(redisClient, cfg.NotifiedTaskSetKey, cfg.NotifiedTaskTTL)
	monitorUsecase := usecase.NewMonitorUsecase(inspector, monitorRepo, mailSender, cfg.AdminNotificationEmail)

	// Setup main loop
	ticker := time.NewTicker(cfg.PollingInterval)
	defer ticker.Stop()
	var wg sync.WaitGroup
	log.Printf("Archive monitor started. Polling interval: %s", cfg.PollingInterval)

	for {
		select {
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Println("Polling for archived tasks...")
				for _, qname := range cfg.QueuesToMonitor {
					if err := monitorUsecase.CheckAndNotifyArchivedTasks(ctx, qname); err != nil {
						log.Printf("ERROR: Failed to process archived tasks for queue '%s': %v", qname, err)
					}
				}
			}()
		case <-ctx.Done():
			log.Println("Shutdown signal received. Waiting for ongoing tasks to complete...")
			wg.Wait()
			log.Println("All tasks completed. Shutting down gracefully.")
			return
		}
	}
}
