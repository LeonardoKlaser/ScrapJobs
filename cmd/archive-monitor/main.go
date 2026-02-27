package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"web-scrapper/logging"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
)

func main() {
	logging.Logger.Info().Msg("Starting Archive Monitor service...")
	cfg, err := utils.LoadMonitorConfig()
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("FATAL: Failed to load configuration")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize dependencies using the loaded config.
	// cfg.RedisAddr may be a redis:// URL (from REDIS_URL) or plain host:port.
	var asynqRedisOpt asynq.RedisConnOpt = asynq.RedisClientOpt{Addr: cfg.RedisAddr}
	if parsed, parseErr := asynq.ParseRedisURI(cfg.RedisAddr); parseErr == nil {
		asynqRedisOpt = parsed
	}
	inspector := asynq.NewInspector(asynqRedisOpt)
	defer inspector.Close()

	monitorUsecase := usecase.NewMonitorUsecase(inspector)

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
