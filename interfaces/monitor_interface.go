package interfaces 

import(
	"context"
	"github.com/hibiken/asynq"
)

type MonitorRepositoryInterface interface{
	FilterNotifiedTasks(ctx context.Context, tasks []*asynq.TaskInfo) ([]*asynq.TaskInfo, error)
	MarkTaskAsNotified(ctx context.Context, taskID string) error
}