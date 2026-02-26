package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/repository"
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
	dashboardRepo  *repository.DashboardRepository
}

func NewTaskProcessor(
	scraper usecase.JobUseCase,
	notifier usecase.NotificationsUsecase,
	paymentUC *usecase.PaymentUsecase,
	emailSvc interfaces.EmailService,
	client *asynq.Client,
	dashboardRepo *repository.DashboardRepository,
) *TaskProcessor {
	return &TaskProcessor{
		_scraper:       scraper,
		_notifier:      notifier,
		paymentUsecase: paymentUC,
		emailService:   emailSvc,
		_client:        client,
		dashboardRepo:  dashboardRepo,
	}
}

func (p *TaskProcessor) HandleScrapeSiteTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ScrapeSitePayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleScrapeSiteTask")
		return fmt.Errorf("error to get payload: %w", err)
	}

	logging.Logger.Info().Int("site_id", payload.SiteID).Msg("Processing task to scrap site")

	_, err := p._scraper.ScrapeAndStoreJobs(ctx, payload.SiteScrapingConfig)
	if err != nil {
		logging.Logger.Warn().Err(err).Int("site_id", payload.SiteID).Msg("ScrapeAndStoreJobs failed but task will not be retried")
		if p.dashboardRepo != nil {
			recErr := p.dashboardRepo.RecordScrapingError(payload.SiteID, payload.SiteScrapingConfig.SiteName, err.Error(), t.ResultWriter().TaskID())
			if recErr != nil {
				logging.Logger.Error().Err(recErr).Msg("Failed to record scraping error")
			}
		}
	}

	logging.Logger.Info().Int("site_id", payload.SiteID).Msg("Scraping task completed")
	return nil
}

func (p *TaskProcessor) HandleMatchUserTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.MatchUserPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleMatchUserTask")
		return fmt.Errorf("error decoding MatchUserPayload: %w", err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Processing match job for user")

	if err := p._notifier.MatchJobsForUser(ctx, payload.UserID); err != nil {
		logging.Logger.Error().Err(err).Int("user_id", payload.UserID).Msg("MatchJobsForUser failed")
		return fmt.Errorf("error matching jobs for user %d: %w", payload.UserID, err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Match job completed for user")
	return nil
}

func (p *TaskProcessor) HandleSendDigestTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.SendDigestPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleSendDigestTask")
		return fmt.Errorf("error decoding SendDigestPayload: %w", err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Processing digest email for user")

	if err := p._notifier.SendDigestForUser(ctx, payload.UserID); err != nil {
		logging.Logger.Error().Err(err).Int("user_id", payload.UserID).Msg("SendDigestForUser failed")
		return fmt.Errorf("error sending digest for user %d: %w", payload.UserID, err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Digest email sent for user")
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
