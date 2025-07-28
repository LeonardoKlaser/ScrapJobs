package repository

import (
	"context"
	"log"
	"time"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type MonitorRepository struct{
	rdb *redis.Client
	setKey string
	ttl time.Duration
}

func NewMonitorRepository(rdb *redis.Client, setKey string, ttl time.Duration) *MonitorRepository{
	return &MonitorRepository{
		rdb: rdb,
		setKey: setKey,
		ttl: ttl,
	}
}

func (r *MonitorRepository) FilterNotifiedTasks(ctx context.Context, tasks []*asynq.TaskInfo) ([]*asynq.TaskInfo, error){
	if len(tasks) == 0 {
		return nil, nil
	}

	taskIDs := make([]interface{}, len(tasks))
	for i, t := range tasks{
		taskIDs[i] = t.ID
	}

	isMember, err := r.rdb.SMIsMember(ctx, r.setKey, taskIDs...).Result()
	if err != nil {
		if err == redis.Nil{
			return tasks, nil
		}
		return nil, err
	}

	var unnotifiedTasks []*asynq.TaskInfo
	for i, member := range isMember{
		if !member {
			unnotifiedTasks = append(unnotifiedTasks, tasks[i])
		}
	}
	return unnotifiedTasks, nil
}

func (r *MonitorRepository) MarkTaskAsNotified(ctx context.Context, taskID string) error {
	pipe := r.rdb.Pipeline()
	pipe.SAdd(ctx, r.setKey, taskID)
	pipe.Expire(ctx, r.setKey, r.ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to mark task %s as notified: %v", taskID, err)
		return err
	}
	return nil
}
