package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"web-scrapper/infra/ses"
	"web-scrapper/logging"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Println("Starting Archive Monitor service...")
	cfg, err := utils.LoadMonitorConfig()
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("FATAL: Failed to load configuration")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize dependencies using the loaded config
	redisOpt, err := redis.ParseURL(cfg.RedisAddr)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("FATAL: could not parse redis address")
	}
	redisClient := redis.NewClient(redisOpt)
	defer redisClient.Close()

	asynqRedisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}
	inspector := asynq.NewInspector(asynqRedisOpt)
	defer inspector.Close()

	awsCfg, err := ses.LoadAWSConfig(ctx)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("FATAL: could not load aws config")
	}
	sesClient := ses.LoadAWSClient(awsCfg)
	mailSender := ses.NewSESMailSender(sesClient, cfg.SenderEmail)

	cloudwatchClient := cloudwatch.NewFromConfig(awsCfg)

	// Instantiate architectural components
	monitorRepo := repository.NewMonitorRepository(redisClient, cfg.NotifiedTaskSetKey, cfg.NotifiedTaskTTL)
	monitorUsecase := usecase.NewMonitorUsecase(inspector, monitorRepo, mailSender, cloudwatchClient, cfg.AdminNotificationEmail)

	// Setup main loop
	ticker := time.NewTicker(cfg.PollingInterval)
	defer ticker.Stop()
	var wg sync.WaitGroup
	log.Printf("Archive monitor started. Polling interval: %s", cfg.PollingInterval)


	runCheck(ctx, &wg, cfg.QueuesToMonitor, monitorUsecase)

	for {
		select {
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				runCheck(ctx, &wg, cfg.QueuesToMonitor, monitorUsecase)
			}()
		case <-ctx.Done():
			log.Println("Shutdown signal received. Waiting for ongoing tasks to complete...")
			wg.Wait()
			log.Println("All tasks completed. Shutting down gracefully.")
			return
		}
	}
}

func runCheck(ctx context.Context, wg *sync.WaitGroup, queues []string, uc *usecase.MonitorUsecase) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Polling for archived tasks...")
		for _, qname := range queues {
			if err := uc.CheckAndNotifyArchivedTasks(ctx, qname); err != nil {
				log.Printf("ERROR: Failed to process archived tasks for queue '%s': %v", qname, err)
			}
		}
	}()
}
