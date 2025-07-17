package repository

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type NotificationRepository struct{
	connection *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository{
	return &NotificationRepository{
		connection: db,
	}
}

func (db *NotificationRepository) InsertNewNotification(jobId int, userId int) error{
	query := `INSERT INTO job_notifications (user_id, job_id) VALUES($1, $2)`

	_, err := db.connection.Exec(query , userId, jobId)

	if err != nil{
		return fmt.Errorf("error to insert new notification to user %d and job %d: %w", userId, jobId, err)
	}

	return nil
}

func (db *NotificationRepository) GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error){
	notified := make(map[int]bool)
    if len(jobs) == 0 {
        return notified, nil
    }

    query := `
        SELECT job_id 
        FROM job_notifications 
        WHERE user_id = $1 AND job_id = ANY($2)`

	rows, err := db.connection.Query(query, userId, pq.Array(jobs) )
	if err != nil {
		return nil, fmt.Errorf("error fetching notified jobs for user %d: %w", userId, err)
    }
	
	defer rows.Close()

	for rows.Next(){
		var jobId int
		if err := rows.Scan(&jobId); err != nil {
			return nil, fmt.Errorf("error scanning notified job ID: %w", err)
		}
		notified[jobId] = true
	}

	return notified, rows.Err()
}