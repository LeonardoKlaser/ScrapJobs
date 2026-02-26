package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"web-scrapper/infra/db"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
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
            logging.Logger.Fatal().Err(err).Msg("Failed to get secrets from AWS Secrets Manager")
        }
	} else {
        secrets = &model.AppSecrets{
            DBHost:     os.Getenv("HOST_DB"),
            DBPort:     os.Getenv("PORT_DB"),
            DBUser: os.Getenv("USER_DB"),
            DBPassword: os.Getenv("PASSWORD_DB"),
            DBName:   os.Getenv("DBNAME"),
            RedisAddr: os.Getenv("REDIS_ADDR"),
        }
    }

    client := asynq.NewClient(asynq.RedisClientOpt{Addr: secrets.RedisAddr})
    defer client.Close()

    dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort,secrets.DBUser,secrets.DBPassword,secrets.DBName)
    if err != nil {
        logging.Logger.Fatal().Err(err).Msg("Scheduler could not connect to db")
    }
    defer dbConnection.Close()
    siteRepo := repository.NewSiteCareerRepository(dbConnection)
    jobRepo := repository.NewJobRepository(dbConnection)
    userSiteRepo := repository.NewUserSiteRepository(dbConnection)
    notificationRepo := repository.NewNotificationRepository(dbConnection)

    // Ticker to run every 120 minutes
    ticker := time.NewTicker(120 * time.Minute)
    defer ticker.Stop()

    tickerDeleteJobs := time.NewTicker(24 * time.Hour)
    defer tickerDeleteJobs.Stop()

    tickerMatch := time.NewTicker(4 * time.Hour)
    defer tickerMatch.Stop()

    tickerDigest := time.NewTicker(8 * time.Hour)
    defer tickerDigest.Stop()

    // Graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    // Run initial scraping on startup
    go enqueueScrapingTasks(context.Background(), siteRepo, client)
    go enqueueMatchTasks(context.Background(), userSiteRepo, client)
    go enqueueDigestTasks(context.Background(), notificationRepo, client)

    for {
        select {
        case <-ticker.C:
            go enqueueScrapingTasks(context.Background(), siteRepo, client)
        case <-tickerDeleteJobs.C:
            go func() {
                if err := jobRepo.DeleteOldJobs(); err != nil {
                    logging.Logger.Error().Err(err).Msg("ERROR: failed to delete old jobs")
                }
            }()
        case <-tickerMatch.C:
            go enqueueMatchTasks(context.Background(), userSiteRepo, client)
        case <-tickerDigest.C:
            go enqueueDigestTasks(context.Background(), notificationRepo, client)
        case sig := <-sigCh:
            logging.Logger.Info().Str("signal", sig.String()).Msg("Scheduler shutting down")
            return
        }
    }
}

func enqueueScrapingTasks(ctx context.Context, siteRepo *repository.SiteCareerRepository, client *asynq.Client) {
    sites, err := siteRepo.GetAllSites()
    if err != nil {
        logging.Logger.Error().Err(err).Msg("Scheduler can't get sites from database")
        return
    }
    logging.Logger.Info().Int("count", len(sites)).Msg("Sites coletados")
    var wg sync.WaitGroup
    for _, site := range sites {
        wg.Add(1)
        go func(s model.SiteScrapingConfig){
            defer wg.Done()
            payload, err := json.Marshal(tasks.ScrapeSitePayload{
                SiteID: s.ID,
                SiteScrapingConfig: s,
            })
            if err != nil {
                logging.Logger.Error().Err(err).Str("site_name", s.SiteName).Msg("Could not marshal task for site")
                return
            }

            task := asynq.NewTask(tasks.TypeScrapSite, payload, asynq.MaxRetry(3))
            info, err := client.EnqueueContext(ctx, task)
            if err != nil {
                logging.Logger.Error().Err(err).Str("site_name", s.SiteName).Msg("Could not enqueue task for site")
            } else {
                logging.Logger.Info().Str("site_name", s.SiteName).Str("task_id", info.ID).Msg("Task enqueued for site")
            }
        }(site)
    }
    wg.Wait()
}

func enqueueMatchTasks(ctx context.Context, userSiteRepo *repository.UserSiteRepository, client *asynq.Client) {
	userIDs, err := userSiteRepo.GetActiveUserIDs()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Scheduler can't get active user IDs")
		return
	}
	logging.Logger.Info().Int("count", len(userIDs)).Msg("Active users for match")

	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()
			payload, err := json.Marshal(tasks.MatchUserPayload{UserID: uid})
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not marshal match task")
				return
			}

			task := asynq.NewTask(tasks.TypeMatchUser, payload, asynq.MaxRetry(3))
			info, err := client.EnqueueContext(ctx, task)
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not enqueue match task")
			} else {
				logging.Logger.Info().Int("user_id", uid).Str("task_id", info.ID).Msg("Match task enqueued")
			}
		}(userID)
	}
	wg.Wait()
}

func enqueueDigestTasks(ctx context.Context, notificationRepo *repository.NotificationRepository, client *asynq.Client) {
	userIDs, err := notificationRepo.GetUserIDsWithPendingNotifications()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Scheduler can't get users with pending notifications")
		return
	}
	logging.Logger.Info().Int("count", len(userIDs)).Msg("Users with pending digest notifications")

	if len(userIDs) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()
			payload, err := json.Marshal(tasks.SendDigestPayload{UserID: uid})
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not marshal digest task")
				return
			}

			task := asynq.NewTask(tasks.TypeSendDigest, payload, asynq.MaxRetry(3))
			info, err := client.EnqueueContext(ctx, task)
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not enqueue digest task")
			} else {
				logging.Logger.Info().Int("user_id", uid).Str("task_id", info.ID).Msg("Digest task enqueued")
			}
		}(userID)
	}
	wg.Wait()
}