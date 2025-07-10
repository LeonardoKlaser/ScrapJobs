package usecase

import (
	"context"
	"encoding/json"
	"log"
	"web-scrapper/repository"
	"web-scrapper/tasks"
	"github.com/hibiken/asynq"
)

type TaskEnqueuer struct {
	_siteRepo repository.SiteCareerRepository
	_client *asynq.Client
}

func NewTaskEnqueuer (siteRepo repository.SiteCareerRepository, client *asynq.Client) *TaskEnqueuer{
	return &TaskEnqueuer{
		_siteRepo: siteRepo,
		_client: client,
	}
}

func (o *TaskEnqueuer) ExecuteScrapingCycle(ctx context.Context){
	sites, err := o._siteRepo.GetAllSites()
	if err != nil {
		log.Printf("ERROR: TaskEnqueuer cant get the sites from database: %v", err)
	}

	for _, site := range sites{
		payload, _ := json.Marshal(tasks.ScrapeSitePayload{
			SiteID: site.ID,
			SiteScrapingConfig: site,
		})

		task := asynq.NewTask(tasks.TypeScrapSite, payload)

		info, err := o._client.EnqueueContext(ctx, task)

		if err != nil {
			log.Printf("ERROR: Error to enqueue task to site %s : %v", site.SiteName, err)
		}else{
			log.Printf("INFO: task to scrape %s enqueued. ID: %s ", site.SiteName, info.ID)
		}
	}
}