package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
	"web-scrapper/infra/db"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"web-scrapper/utils"
	"web-scrapper/middleware"
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
            middleware.Logger.Fatal().Err(err).Msg("Failed to get secrets from AWS Secrets Manager")
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
        middleware.Logger.Fatal().Err(err).Msg("Scheduler could not connect to db")
    }
    siteRepo := repository.NewSiteCareerRepository(dbConnection)
    jobRepo := repository.NewJobRepository(dbConnection)

    // Ticker to run hourly
    ticker := time.NewTicker(60 * time.Minute)
    defer ticker.Stop()

    tickerDeleteJobs := time.NewTicker(24 * time.Hour)
    defer tickerDeleteJobs.Stop()

    for {
        select{
        case <-ticker.C:
            go enqueueScrapingTasks(context.Background(), siteRepo, client)
        case <-tickerDeleteJobs.C:
            go func(){
                if err := jobRepo.DeleteOldJobs(); err != nil{
                    middleware.Logger.Fatal().Err(err).Msg("ERROR: failed to delete old jobs")
                }
            }()
        }
    }

}

func enqueueScrapingTasks(ctx context.Context, siteRepo *repository.SiteCareerRepository, client *asynq.Client) {
    sites, err := siteRepo.GetAllSites()
    if err != nil {
        log.Printf("ERROR: Scheduler can't get sites from database: %v", err)
        return
    }

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
                log.Printf("ERROR: Could not marshal task for site %s: %v", s.SiteName, err)
                return
            }

            task := asynq.NewTask(tasks.TypeScrapSite, payload, asynq.MaxRetry(3))
            info, err := client.EnqueueContext(ctx, task)
            if err != nil {
                log.Printf("ERROR: Could not enqueue task for site %s: %v", s.SiteName, err)
            } else {
                log.Printf("INFO: Task enqueued for site %s. ID: %s", s.SiteName, info.ID)
            }
        }(site)
    }
    wg.Wait()
}