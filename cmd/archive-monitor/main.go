package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"web-scrapper/infra/ses"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisAddr              string
	PollingInterval        time.Duration
	AdminNotificationEmail string
	NotifiedTaskSetKey     string
	NotifiedTaskTTL        time.Duration
	QueuesToMonitor        []string
	SenderEmail            string
}

func loadConfig() (*Config, error ){
	if os.Getenv("GIN_MODE") != "release"{
		godotenv.Load()
	}

	pollingIntervalStr := os.Getenv("MONITOR_POLLING_INTERVAL")
	if pollingIntervalStr == "" {
		pollingIntervalStr = "5m" // Default
	}
	pollingInterval, err := time.ParseDuration(pollingIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MONITOR_POLLING_INTERVAL: %w", err)
	}

	notifiedTaskTTLStr := os.Getenv("NOTIFIED_TASK_TTL")
	if notifiedTaskTTLStr == "" {
		notifiedTaskTTLStr = "168h" // Default 7 days
	}
	notifiedTaskTTL, err := time.ParseDuration(notifiedTaskTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid NOTIFIED_TASK_TTL: %w", err)
	}

	queuesStr := os.Getenv("QUEUES_TO_MONITOR")
	if queuesStr == "" {
		queuesStr = "critical,default,low"
	}

	return &Config{
		RedisAddr:              os.Getenv("REDIS_ADDR"),
		PollingInterval:        pollingInterval,
		AdminNotificationEmail: os.Getenv("ADMIN_NOTIFICATION_EMAIL"),
		NotifiedTaskSetKey:     os.Getenv("NOTIFIED_TASK_SET_KEY"),
		NotifiedTaskTTL:        notifiedTaskTTL,
		QueuesToMonitor:        strings.Split(queuesStr, ","),
		SenderEmail:            os.Getenv("SENDER_EMAIL"),
	}, nil
}

func main(){
	log.Println("starting Archive Monitor Service...")
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("FATAL: failed to load configuration: %v",err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisOpt, err := redis.ParseURL(cfg.RedisAddr)
	if err != nil{
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

	monitorRepo := repository.NewMonitorRepository(redisClient, cfg.NotifiedTaskSetKey, cfg.NotifiedTaskTTL)
	monitorUseCase := usecase.NewMonitorUsecase(inspector, monitorRepo, mailSender, cfg.AdminNotificationEmail)

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
					if err := monitorUseCase.CheckAndNotifyArchivedTasks(ctx, qname); err != nil {
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