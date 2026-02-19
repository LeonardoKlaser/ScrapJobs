package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"web-scrapper/infra/ses"
	"web-scrapper/logging"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func main() {
	logging.Logger.Info().Msg("Starting Archive Monitor service...")
	cfg, err := utils.LoadMonitorConfig()
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("FATAL: Failed to load configuration")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize dependencies using the loaded config
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	defer redisClient.Close()

	asynqRedisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}
	inspector := asynq.NewInspector(asynqRedisOpt)
	defer inspector.Close()

	awsCfg, err := ses.LoadAWSConfig(ctx)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("FATAL: could not load aws config")
	}

	cloudwatchClient := cloudwatch.NewFromConfig(awsCfg)

	monitorUsecase := usecase.NewMonitorUsecase(inspector, cloudwatchClient)

	// Setup main loop
	ticker := time.NewTicker(cfg.PollingInterval)
	defer ticker.Stop()
	var wg sync.WaitGroup
	logging.Logger.Info().Dur("polling_interval", cfg.PollingInterval).Msg("Archive monitor started")

	wg.Add(1)
	go func() {
		defer wg.Done()
		runCheck(ctx, cfg.QueuesToMonitor, monitorUsecase)
	}()

	for {
		select {
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				runCheck(ctx, cfg.QueuesToMonitor, monitorUsecase)
			}()
		case <-ctx.Done():
			logging.Logger.Info().Msg("Shutdown signal received. Waiting for ongoing tasks to complete...")
			wg.Wait()
			logging.Logger.Info().Msg("All tasks completed. Shutting down gracefully.")
			return
		}
	}
}

func runCheck(ctx context.Context, queues []string, uc *usecase.MonitorUsecase) {
	logging.Logger.Info().Msg("Polling for archived tasks...")
	for _, qname := range queues {
		if err := uc.CheckAndNotifyArchivedTasks(ctx, qname); err != nil {
			logging.Logger.Error().Err(err).Str("queue", qname).Msg("Failed to process archived tasks")
		}
	}
}
