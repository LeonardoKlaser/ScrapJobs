package usecase

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestCheckAndNotifyArchivedTasks_QueueNotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer inspector.Close()

	uc := NewMonitorUsecase(inspector)

	// Queue doesn't exist — should not return error (handled gracefully)
	err := uc.CheckAndNotifyArchivedTasks(context.Background(), "nonexistent_queue")
	assert.NoError(t, err)
}

func TestCheckAndNotifyArchivedTasks_EmptyQueue(t *testing.T) {
	mr := miniredis.RunT(t)

	redisOpt := asynq.RedisClientOpt{Addr: mr.Addr()}
	client := asynq.NewClient(redisOpt)
	// Enqueue a task to create the queue structure
	_, err := client.Enqueue(asynq.NewTask("test:task", nil), asynq.Queue("test_queue"))
	assert.NoError(t, err)
	client.Close()

	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()

	uc := NewMonitorUsecase(inspector)

	// Queue exists but no archived tasks — should succeed
	err = uc.CheckAndNotifyArchivedTasks(context.Background(), "test_queue")
	assert.NoError(t, err)
}

func TestPublishArchivedTasksMetric_NoError(t *testing.T) {
	mr := miniredis.RunT(t)
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer inspector.Close()

	uc := NewMonitorUsecase(inspector)

	// publishArchivedTasksMetric now just logs, should always return nil
	err := uc.publishArchivedTasksMetric(context.Background(), "default", 5)
	assert.NoError(t, err)

	err = uc.publishArchivedTasksMetric(context.Background(), "default", 0)
	assert.NoError(t, err)
}
