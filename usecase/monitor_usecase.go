package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"web-scrapper/infra/ses"
	"web-scrapper/interfaces"

	"github.com/hibiken/asynq"
)

type MonitorUsecase struct{
	inspector *asynq.Inspector
	monitorRepo interfaces.MonitorRepositoryInterface
	emailSvc ses.SESMailSender
	adminEmail string
}

func NewMonitorUsecase(
	inspector *asynq.Inspector,
	monitorRepo interfaces.MonitorRepositoryInterface,
	emailSvc *ses.SESMailSender,
	adminEmail string,
) *MonitorUsecase {
	return &MonitorUsecase{
		inspector:   inspector,
		monitorRepo: monitorRepo,
		emailSvc:    emailSvc,
		adminEmail:  adminEmail,
	}
}

func (uc *MonitorUsecase) CheckAndNotifyArchivedTasks(ctx context.Context, queueName string) error{
	const pageSize = 100
	pageNum := 1
	for {
		tasks, err := uc.inspector.ListArchivedTasks(queueName, asynq.Page(pageNum), asynq.PageSize(pageSize))
		if err != nil {
			return fmt.Errorf("could not list archived tasks: %w", err)
		}
		if len(tasks) == 0{
			break	
		}

		unnotifiedTasks, err := uc.monitorRepo.FilterNotifiedTasks(ctx, tasks)
		if err != nil {
			log.Printf("ERROR: Could not filter notified tasks for queue '%s': %v", queueName, err)
			continue
		}

		for _, task := range unnotifiedTasks{
			if err := uc.sendNotification(ctx, task); err != nil {
				log.Printf("ERROR: Failed to send notification for task %s: %v", task.ID, err)
			}else {
				if err := uc.monitorRepo.MarkTaskAsNotified(ctx, task.ID); err != nil {
					log.Printf("ERROR: Failed to mark task %s as notified after sending email: %v", task.ID, err)
				}
			}
		}

		pageNum++
	}

	return nil
}


func (uc *MonitorUsecase) sendNotification(ctx context.Context, task *asynq.TaskInfo) error {
	if uc.adminEmail == "" {
		return fmt.Errorf("admin notification email is not configured")
	}

	subject := fmt.Sprintf("Tarefa Falhou Permanentemente no ScrapJobs: %s", task.Type)

	var prettyPayload string
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(task.Payload, &payloadMap); err == nil {
		if prettyBytes, err := json.MarshalIndent(payloadMap, "", "  "); err == nil {
			prettyPayload = string(prettyBytes)
		}
	} else {
		prettyPayload = string(task.Payload)
	}

	body := fmt.Sprintf(`
        <h1>Alerta de Falha de Tarefa</h1>
        <p>Uma tarefa falhou em todas as suas tentativas e foi movida para o estado <strong>Archived</strong>.</p>
        <h3>Detalhes da Tarefa</h3>
        <ul>
            <li><strong>ID da Tarefa:</strong> %s</li>
            <li><strong>Tipo:</strong> %s</li>
            <li><strong>Fila:</strong> %s</li>
            <li><strong>Última Falha:</strong> %s</li>
        </ul>
        <h3>Causa da Falha</h3>
        <p>A mensagem do último erro foi:</p>
        <pre>%s</pre>
        <h3>Payload da Tarefa</h3>
        <pre>%s</pre>
        <h3>Ação Sugerida</h3>
        <p>Investigue a causa da falha. Use a CLI do Asynq ou o AsynqMon para inspecionar, reenfileirar (run) ou excluir (delete) a tarefa, se necessário.</p>
    `, task.ID, task.Type, task.Queue, task.LastFailedAt.Format(time.RFC1123Z), task.LastErr, prettyPayload)

	return uc.emailSvc.SendEmail(ctx, uc.adminEmail, subject, body, body)
}