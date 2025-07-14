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

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
    godotenv.Load()
    log.Println("Scheduler starting...")
    
    var redisAddr = os.Getenv("REDIS_ADDR")
    client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
    defer client.Close()

    dbConnection, err := db.ConnectDB(os.Getenv("HOST_DB"), os.Getenv("PORT_DB"), os.Getenv("USER_DB"), os.Getenv("PASSWORD_DB"), os.Getenv("DBNAME"))
    if err != nil {
        log.Fatalf("Scheduler could not connect to db: %v", err)
    }
    siteRepo := repository.NewSiteCareerRepository(dbConnection)
    jobRepo := repository.NewJobRepository(dbConnection)

    // Ticker to run hourly
    ticker := time.NewTicker(10 * time.Minute)
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
                    log.Printf("ERROR: failed to delete old jobs: %v", err)
                }
            }()
        }
    }

}

func enqueueScrapingTasks(ctx context.Context, siteRepo repository.SiteCareerRepository, client *asynq.Client) {
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

            task := asynq.NewTask(tasks.TypeScrapSite, payload)
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