package usecase

import (
	"context"
	"strings"
	"web-scrapper/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hibiken/asynq"
)

type MonitorUsecase struct{
	inspector *asynq.Inspector
	cloudWatchSvc *cloudwatch.Client
}

func NewMonitorUsecase(
	inspector *asynq.Inspector,
	cloudWatchSvc *cloudwatch.Client,
) *MonitorUsecase {
	return &MonitorUsecase{
		inspector:   inspector,
		cloudWatchSvc: cloudWatchSvc,
	}
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
		return uc.publishArchivedTasksMetric(ctx, queueName, 0)
	}

	archivedCount := 0
	if queueInfo != nil {
		archivedCount = queueInfo.Archived
	}

	return uc.publishArchivedTasksMetric(ctx, queueName, archivedCount); 
}

func (uc *MonitorUsecase) publishArchivedTasksMetric(ctx context.Context, queueName string, count int) error {
	_, err := uc.cloudWatchSvc.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("ScrapJobs/Application"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("AsynqArchivedQueueDepth"),
				Value:      aws.Float64(float64(count)),
				Unit:       types.StandardUnitCount,
				Dimensions: []types.Dimension{ // Adicionar a fila como dimensão é uma boa prática
					{
						Name:  aws.String("QueueName"),
						Value: aws.String(queueName),
					},
				},
			},
		},
	})
	if err != nil {
        logging.Logger.Error().Err(err).Str("queue", queueName).Msg("Failed to publish CloudWatch metric")
    } else {
        logging.Logger.Info().Str("queue", queueName).Int("archived_count", count).Msg("Successfully published archived tasks metric")
    }
	return err
}
