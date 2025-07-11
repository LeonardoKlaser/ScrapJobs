package main

import (
	"context"
	"log"
	"time"
	"web-scrapper/tasks" // Importa a definição da sua tarefa
    "web-scrapper/repository" // Para buscar os sites
    "web-scrapper/infra/db" // Para conectar ao banco
    "encoding/json"
    "os"

	"github.com/hibiken/asynq"
    "github.com/joho/godotenv"
)

const redisAddr = "redis:6379"

func main() {
    godotenv.Load()
    log.Println("Scheduler starting...")

    client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
    defer client.Close()

    dbConnection, err := db.ConnectDB(os.Getenv("HOST_DB"), os.Getenv("PORT_DB"), os.Getenv("USER_DB"), os.Getenv("PASSWORD_DB"), os.Getenv("DBNAME"))
    if err != nil {
        log.Fatalf("Scheduler could not connect to db: %v", err)
    }
    siteRepo := repository.NewSiteCareerRepository(dbConnection)

    // Ticker to run hourly
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for {
        enqueueScrapingTasks(context.Background(), siteRepo, client)

        log.Println("Tasks enqueued. Scheduler is now waiting for the next hour.")

        <-ticker.C
    }
}

func enqueueScrapingTasks(ctx context.Context, siteRepo repository.SiteCareerRepository, client *asynq.Client) {
    sites, err := siteRepo.GetAllSites()
    if err != nil {
        log.Printf("ERROR: Scheduler can't get sites from database: %v", err)
        return
    }

    for _, site := range sites {
        payload, _ := json.Marshal(tasks.ScrapeSitePayload{
            SiteID:             site.ID,
            SiteScrapingConfig: site,
        })

        task := asynq.NewTask(tasks.TypeScrapSite, payload)
        info, err := client.EnqueueContext(ctx, task)
        if err != nil {
            log.Printf("ERROR: Could not enqueue task for site %s: %v", site.SiteName, err)
        } else {
            log.Printf("INFO: Task enqueued for site %s. ID: %s", site.SiteName, info.ID)
        }
    }
}