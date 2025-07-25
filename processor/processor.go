package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"web-scrapper/tasks"
	"web-scrapper/usecase"
	"web-scrapper/infra/ses"
	"github.com/hibiken/asynq"
)

type TaskProcessor struct{
	_scraper usecase.JobUseCase
	_notifier usecase.NotificationsUsecase
	_client *asynq.Client
	_email *ses.SESMailSender
}

func NewTaskProcessor(scraper usecase.JobUseCase, notifier usecase.NotificationsUsecase, client *asynq.Client, email *ses.SESMailSender) *TaskProcessor{
	return &TaskProcessor{_scraper: scraper, _notifier: notifier, _client: client, _email: email}
}

func (p *TaskProcessor) HandleScrapeSiteTask(ctx context.Context, t *asynq.Task) error{
	var payload tasks.ScrapeSitePayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("error to get payload: %w", err)
	}

	log.Printf("INFO Processing task to scrap siteID: %d", payload.SiteID)

	newJobs, err := p._scraper.ScrapeAndStoreJobs(ctx, payload.SiteScrapingConfig)
	if err != nil {
		log.Printf("WARN: ScrapeAndStoreJobs for site %d failed but task will not be retried. Error: %v", payload.SiteID, err)
	}

	if len(newJobs) == 0 {
		log.Printf("INFO: no new jobs for site id: %d", payload.SiteID)
		return nil
	}

	resultsPayload, _ := json.Marshal(tasks.ProcessResultsPayload{
		SiteID: payload.SiteID,
		Jobs: newJobs,
	})

	nextTask := asynq.NewTask(tasks.TypeProcessResults, resultsPayload, asynq.MaxRetry(3))
	info, err := p._client.EnqueueContext(ctx, nextTask)
	if err != nil {
		log.Printf("error to enqueue site: %d result task : %v", payload.SiteID,err)
		return nil
	}

	log.Printf("INFO: siteID: %d scrap finished. Process task enqueued: %s", payload.SiteID, info.ID)
	return nil
}

func (p *TaskProcessor) HandleProcessResultsTask(ctx context.Context, t *asynq.Task) error {
    var payload tasks.ProcessResultsPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("error to get payload results: %w", err)
    }

    log.Printf("INFO: processing result to site: %d, jobs: %d", payload.SiteID, len(payload.Jobs))
    
    
    err := p._notifier.FindMatchesAndNotify(payload.SiteID, payload.Jobs)
    if err != nil {
        return fmt.Errorf("error to notify to site %d: %w", payload.SiteID, err)
    }
    
    log.Printf("INFO: result process to site %d finished.", payload.SiteID)
	return nil
}

func (p *TaskProcessor) HandleDeadQueueLetter(ctx context.Context, t *asynq.Task){
	retryCount, _ := asynq.GetRetryCount(ctx)
	
	log.Printf("ALERT: Received a task in Dead-Letter Queue. TaskID: %s, Type: %s", t.ResultWriter().TaskID(), t.Type())

	subject := fmt.Sprintf("[ALERTA SCRAPJOBS] Tarefa falhou: %s", t.Type())
	body := fmt.Sprintf(`
		Uma tarefa falhou permanentemente:
		Detalhes da tarefa:
		ID:%s
		tipo da tarefa: %s
		Fila: dead
		Numero de retentativas: %d

		Payload da tarefa:
		<pre>%s</pre>


		Investiga isso aí
	`, t.ResultWriter().TaskID(), t.Type(), retryCount, string(t.Payload()))

	adminEmail := "admin@scrapjobs.com.br"

	err := p._email.SendEmail(ctx, adminEmail, subject, body, body )
	if err != nil {
		log.Printf("FATAL: Could not send DLQ alert email for TaskID %s. Error: %v", t.ResultWriter().TaskID(), err)
	} else {
		log.Printf("Admin alert email sent successfully for TaskID: %s", t.ResultWriter().TaskID())
	}
}