package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"database/sql"
	_ "time/tzdata"
	"web-scrapper/infra/db"
	"web-scrapper/logging"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"web-scrapper/utils"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const (
	matchUniqueTTL  = 3*time.Hour + 50*time.Minute
	digestUniqueTTL = 7*time.Hour + 50*time.Minute
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

    if err := utils.ValidateSecrets(secrets); err != nil {
        logging.Logger.Fatal().Err(err).Msg("Invalid configuration")
    }

    redisAddr := secrets.RedisAddr
    if redisAddr == "" {
        redisAddr = os.Getenv("REDIS_URL")
    }

    var asynqRedisOpt asynq.RedisConnOpt = asynq.RedisClientOpt{Addr: redisAddr}
    if parsed, parseErr := asynq.ParseRedisURI(redisAddr); parseErr == nil {
        asynqRedisOpt = parsed
    }

    client := asynq.NewClient(asynqRedisOpt)
    defer client.Close()

    var dbConnection *sql.DB
    if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
        dbConnection, err = db.ConnectDBFromURL(dbURL, func(d *sql.DB) {
            d.SetMaxOpenConns(5)
            d.SetMaxIdleConns(2)
        })
    } else {
        dbConnection, err = db.ConnectDB(secrets.DBHost, secrets.DBPort,secrets.DBUser,secrets.DBPassword,secrets.DBName, func(d *sql.DB) {
            d.SetMaxOpenConns(5)
            d.SetMaxIdleConns(2)
        })
    }
    if err != nil {
        logging.Logger.Fatal().Err(err).Msg("Scheduler could not connect to db")
    }
    defer dbConnection.Close()
    siteRepo := repository.NewSiteCareerRepository(dbConnection)
    jobRepo := repository.NewJobRepository(dbConnection)
    userSiteRepo := repository.NewUserSiteRepository(dbConnection)
    notificationRepo := repository.NewNotificationRepository(dbConnection)
    resetRepo := repository.NewPasswordResetRepository(dbConnection)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Failed to load timezone America/Sao_Paulo")
	}

	c := cron.New(cron.WithLocation(loc))

	// Scraping: 7h, 9h, 11h, 13h, 15h, 17h (every 2h from 7-17)
	c.AddFunc("0 7,9,11,13,15,17 * * *", func() {
		enqueueScrapingTasks(ctx, siteRepo, client)
	})

	// Match: 8h and 16h
	c.AddFunc("0 8,16 * * *", func() {
		enqueueMatchTasks(ctx, userSiteRepo, client)
	})

	// Digest/Email: 9h and 17h
	c.AddFunc("0 9,17 * * *", func() {
		enqueueDigestTasks(ctx, notificationRepo, client)
	})

	// Cleanup: 3h daily
	c.AddFunc("0 3 * * *", func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := jobRepo.DeleteOldJobs(); err != nil {
				logging.Logger.Error().Err(err).Msg("ERROR: failed to delete old jobs")
			}
		}()
		go func() {
			defer wg.Done()
			deleted, err := resetRepo.DeleteExpiredTokens()
			if err != nil {
				logging.Logger.Error().Err(err).Msg("ERROR: failed to delete expired reset tokens")
			} else if deleted > 0 {
				logging.Logger.Info().Int64("count", deleted).Msg("Expired reset tokens cleaned up")
			}
		}()
		wg.Wait()
	})

	c.Start()
	defer c.Stop()

	for _, entry := range c.Entries() {
		logging.Logger.Info().Time("next_run", entry.Next).Msg("Scheduled cron entry")
	}
	logging.Logger.Info().Msg("Scheduler started with cron expressions (America/Sao_Paulo)")

	// Block until signal
	<-sigCh
	logging.Logger.Info().Str("signal", "received").Msg("Scheduler shutting down...")
	cancel()
}

func enqueueScrapingTasks(ctx context.Context, siteRepo *repository.SiteCareerRepository, client *asynq.Client) {
    sites, err := siteRepo.GetAllSites()
    if err != nil {
        logging.Logger.Error().Err(err).Msg("Scheduler can't get sites from database")
        return
    }
    logging.Logger.Info().Int("count", len(sites)).Msg("Sites coletados")
    sem := make(chan struct{}, 10)
    var wg sync.WaitGroup
    for _, site := range sites {
        wg.Add(1)
        sem <- struct{}{}
        go func(s model.SiteScrapingConfig){
            defer wg.Done()
            defer func() { <-sem }()
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

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(uid int) {
			defer wg.Done()
			defer func() { <-sem }()
			payload, err := json.Marshal(tasks.MatchUserPayload{UserID: uid})
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not marshal match task")
				return
			}

			task := asynq.NewTask(tasks.TypeMatchUser, payload, asynq.MaxRetry(3), asynq.Unique(matchUniqueTTL))
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

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(uid int) {
			defer wg.Done()
			defer func() { <-sem }()
			payload, err := json.Marshal(tasks.SendDigestPayload{UserID: uid})
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not marshal digest task")
				return
			}

			task := asynq.NewTask(tasks.TypeSendDigest, payload, asynq.MaxRetry(3), asynq.Unique(digestUniqueTTL))
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
