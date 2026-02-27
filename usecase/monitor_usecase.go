package usecase

import (
	"context"
	"strings"
	"web-scrapper/logging"

	"github.com/hibiken/asynq"
)

type MonitorUsecase struct{
	inspector *asynq.Inspector
}

func NewMonitorUsecase(inspector *asynq.Inspector) *MonitorUsecase {
	return &MonitorUsecase{inspector: inspector}
}

func (uc *MonitorUsecase) CheckAndNotifyArchivedTasks(ctx context.Context, queueName string) error{
	queueInfo, err := uc.inspector.GetQueueInfo(queueName)
	if err != nil {
		// --- A CORREÇÃO ESTÁ AQUI ---
		// Verificamos se a mensagem de erro contém o texto específico "does not exist".
		// Esta é a forma robusta de detetar este erro em particular.
		if strings.Contains(err.Error(), "does not exist") {
			logging.Logger.Debug().Str("queue", queueName).Msg("Queue not found, publishing metric with value 0")
			return uc.publishArchivedTasksMetric(ctx, queueName, 0)
		}
		// Para qualquer outro erro, registamos como um erro real.
		logging.Logger.Error().Err(err).Str("queue", queueName).Msg("Could not get queue info")
		return err
	}

	archivedCount := 0
	if queueInfo != nil {
		archivedCount = queueInfo.Archived
	}

	return uc.publishArchivedTasksMetric(ctx, queueName, archivedCount)
}

func (uc *MonitorUsecase) publishArchivedTasksMetric(ctx context.Context, queueName string, count int) error {
	if count > 0 {
		logging.Logger.Warn().
			Str("queue", queueName).
			Int("archived_count", count).
			Str("metric", "AsynqArchivedQueueDepth").
			Msg("Archived tasks detected")
	} else {
		logging.Logger.Info().
			Str("queue", queueName).
			Int("archived_count", 0).
			Str("metric", "AsynqArchivedQueueDepth").
			Msg("No archived tasks")
	}
	return nil
}
