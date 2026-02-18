package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/tasks"
	"web-scrapper/usecase"

	"github.com/hibiken/asynq"
)

type TaskProcessor struct {
	_scraper       usecase.JobUseCase
	_notifier      usecase.NotificationsUsecase
	paymentUsecase *usecase.PaymentUsecase
	emailService   interfaces.EmailService
	_client        *asynq.Client
}

func NewTaskProcessor(
	scraper usecase.JobUseCase,
	notifier usecase.NotificationsUsecase,
	paymentUC *usecase.PaymentUsecase,
	emailSvc interfaces.EmailService,
	client *asynq.Client,
) *TaskProcessor {
	return &TaskProcessor{
		_scraper:       scraper,
		_notifier:      notifier,
		paymentUsecase: paymentUC,
		emailService:   emailSvc,
		_client:        client,
	}
}

func (p *TaskProcessor) HandleScrapeSiteTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ScrapeSitePayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleScrapeSiteTask")
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
		Jobs:   newJobs,
	})

	nextTask := asynq.NewTask(tasks.TypeProcessResults, resultsPayload, asynq.MaxRetry(3))
	info, err := p._client.EnqueueContext(ctx, nextTask)
	if err != nil {
		log.Printf("error to enqueue site: %d result task : %v", payload.SiteID, err)
		return nil
	}

	logging.Logger.Info().Int("site_id", payload.SiteID).Str("next_task_id", info.ID).Msg("Task de processamento de resultados enfileirada")
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

	for _, analysisPayload := range payloadsToEnqueue {
		p_bytes, err := json.Marshal(analysisPayload)
		if err != nil {
			logging.Logger.Error().Err(err).Msg("Failed to marshal analysis payload, skipping task")
			continue
		}

		analysisTask := asynq.NewTask(tasks.TypeAnalyzeUserJob, p_bytes, asynq.MaxRetry(3))
		_, err = p._client.Enqueue(analysisTask, asynq.Queue("default"))
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
		return fmt.Errorf("error to analyze job: %s to user: %d: %w", payload.Job.Title, payload.User.UserId, err)
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
		logging.Logger.Error().Err(err).Int("user_id", payload.User.UserId).Int("job_id", payload.Job.ID).Msg("Erro ao processar notificação única (envio de email ou registro DB)")
		return fmt.Errorf("error processing notification for job %d, user %d: %w", payload.Job.ID, payload.User.UserId, err)
	}

	logging.Logger.Info().Int("user_id", payload.User.UserId).Int("job_id", payload.Job.ID).Msg("Notificação processada com sucesso")
	return nil
}

// HandleCompleteRegistrationTask processa o registro do usuário após confirmação de pagamento.
func (p *TaskProcessor) HandleCompleteRegistrationTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.CompleteRegistrationPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleCompleteRegistrationTask")
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	logging.Logger.Info().Str("pending_reg_id", payload.PendingRegistrationID).Msg("Processando task para completar registro")

	newUser, err := p.paymentUsecase.CompleteRegistration(ctx, payload.PendingRegistrationID)
	if err != nil {
		logging.Logger.Error().Err(err).Str("pending_reg_id", payload.PendingRegistrationID).Msg("Falha ao completar registro no usecase")
		return fmt.Errorf("failed to complete registration: %w", err)
	}

	logging.Logger.Info().Int("user_id", newUser.Id).Str("email", newUser.Email).Msg("Usuário criado com sucesso via task")

	dashboardLink := os.Getenv("FRONTEND_URL") + "/dashboard"
	err = p.emailService.SendWelcomeEmail(ctx, newUser.Email, newUser.Name, dashboardLink)
	if err != nil {
		logging.Logger.Error().Err(err).Int("user_id", newUser.Id).Msg("Falha ao enviar e-mail de boas-vindas após registro")
	} else {
		logging.Logger.Info().Int("user_id", newUser.Id).Msg("E-mail de boas-vindas enviado")
	}

	logging.Logger.Info().Str("pending_reg_id", payload.PendingRegistrationID).Msg("Task de completar registro concluída com sucesso")
	return nil
}
