package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"web-scrapper/infra/ses"
	"web-scrapper/logging"
	"web-scrapper/tasks"
	"web-scrapper/usecase"

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

func (p *TaskProcessor) HandleFindMatchesTask(ctx context.Context, t *asynq.Task) error {
    var payload tasks.ProcessResultsPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("error to get payload results: %w", err)
    }

    log.Printf("INFO: processing result to site: %d, jobs: %d", payload.SiteID, len(payload.Jobs))
    
    
    payloadsToEnqueue, err := p._notifier.FindMatches(payload.SiteID, payload.Jobs)
    if err != nil {
        return fmt.Errorf("error to notify to site %d: %w", payload.SiteID, err)
    }

	for _, analysisPayload := range payloadsToEnqueue{
		p_bytes, err := json.Marshal(analysisPayload)
        if err != nil {
            logging.Logger.Error().Err(err).Msg("Failed to marshal analysis payload, skipping task")
            continue
        }

		analysisTask := asynq.NewTask(tasks.TypeAnalyzeUserJob, p_bytes, asynq.MaxRetry(3))
        _, err = p._client.Enqueue(analysisTask, asynq.Queue("critical"))
        if err != nil {
            logging.Logger.Error().Err(err).Msg("Failed to enqueue analysis task")
        }
	}
    
    log.Printf("INFO: result process to site %d finished.", payload.SiteID)
	return nil
}

func (p *TaskProcessor) HandleAnalyzeJobUserTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.AnalyzeUserJobPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("error to get payload to analysis: %w", err)
	}

	jobAnalyzed, err := p._notifier.ProcessingJobAnalyze(ctx, *payload.Job, payload.User)
	if err != nil {
		return fmt.Errorf("error to analyze job: %s to user: %d: %w", payload.Job.Title, payload.User.UserId ,err)
	}

	payloadJobAnalyzed, err := json.Marshal(jobAnalyzed)
	if err != nil {
        logging.Logger.Error().Err(err).Msg("Failed to marshal notify user payload")
		return nil 
    }

	analyzeTask := asynq.NewTask(tasks.TypeNotifyUser, payloadJobAnalyzed, asynq.MaxRetry(3))

    _, err = p._client.Enqueue(analyzeTask, asynq.Queue("default")) 
    if err != nil {
        logging.Logger.Error().Err(err).Int("user_id", jobAnalyzed.User.UserId).Int("job_id", jobAnalyzed.Job.ID).Msg("Failed to enqueue notification task")
		return err
    }
	return nil
}

func (p *TaskProcessor) HandleNotifyTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.NotifyUserPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("error to get payload to notify: %w", err)
	}

	err := p._notifier.ProcessingSingleNotification(ctx, *payload.Job, payload.User, payload.Analysis)
	if err != nil {
		return fmt.Errorf("error to send notification job: %s to user: %d : %w", payload.Job.Title, payload.User.UserId, err)
	}
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

