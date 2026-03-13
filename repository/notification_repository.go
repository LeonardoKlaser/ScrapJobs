package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"web-scrapper/model"

	"github.com/lib/pq"
)

type NotificationRepository struct {
	connection *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		connection: db,
	}
}

func (db *NotificationRepository) InsertNewNotification(jobId int, userId int) error {
	query := `INSERT INTO job_notifications (user_id, job_id) VALUES($1, $2)`

	_, err := db.connection.Exec(query, userId, jobId)
	if err != nil {
		return fmt.Errorf("error to insert new notification to user %d and job %d: %w", userId, jobId, err)
	}

	return nil
}

func (db *NotificationRepository) GetNotifiedJobIDsForUser(userId int, jobs []int) (map[int]bool, error) {
	notified := make(map[int]bool)
	if len(jobs) == 0 {
		return notified, nil
	}

	query := `
        SELECT job_id 
        FROM job_notifications 
        WHERE user_id = $1 AND job_id = ANY($2)`

	rows, err := db.connection.Query(query, userId, pq.Array(jobs))
	if err != nil {
		return nil, fmt.Errorf("error fetching notified jobs for user %d: %w", userId, err)
	}
	defer rows.Close()

	for rows.Next() {
		var jobId int
		if err := rows.Scan(&jobId); err != nil {
			return nil, fmt.Errorf("error scanning notified job ID: %w", err)
		}
		notified[jobId] = true
	}

	return notified, rows.Err()
}

// GetMonthlyAnalysisCount retorna a quantidade de análises de IA feitas no mês corrente para o usuário
func (db *NotificationRepository) GetMonthlyAnalysisCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM job_notifications WHERE user_id = $1 AND analysis_result IS NOT NULL AND notified_at >= date_trunc('month', CURRENT_DATE)`

	var count int
	err := db.connection.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting monthly analyses for user %d: %w", userID, err)
	}
	return count, nil
}

// GetNotificationsByUser retorna o histórico de notificações de um usuário com dados da vaga
func (db *NotificationRepository) GetNotificationsByUser(userId int, limit int) ([]model.NotificationWithJob, error) {
	query := `
        SELECT 
            jn.id,
            jn.job_id,
            jn.user_id,
            jn.notified_at,
            j.title,
            j.company,
            j.location,
            j.job_link
        FROM job_notifications jn
        INNER JOIN jobs j ON jn.job_id = j.id
        WHERE jn.user_id = $1
        ORDER BY jn.notified_at DESC
        LIMIT $2`

	rows, err := db.connection.Query(query, userId, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching notifications for user %d: %w", userId, err)
	}
	defer rows.Close()

	var notifications []model.NotificationWithJob
	for rows.Next() {
		var n model.NotificationWithJob
		if err := rows.Scan(
			&n.ID,
			&n.JobID,
			&n.UserID,
			&n.NotifiedAt,
			&n.JobTitle,
			&n.JobCompany,
			&n.JobLocation,
			&n.JobLink,
		); err != nil {
			return nil, fmt.Errorf("error scanning notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (db *NotificationRepository) GetUnnotifiedJobsForUser(userID int) ([]model.JobWithFilters, error) {
	query := `
		SELECT j.id, j.title, j.location, j.company, j.job_link, us.filters
		FROM jobs j
		INNER JOIN user_sites us ON j.site_id = us.site_id AND us.user_id = $1
		WHERE j.last_seen_at >= NOW() - INTERVAL '24 hours'
		  AND NOT EXISTS (
			  SELECT 1 FROM job_notifications jn
			  WHERE jn.user_id = $1 AND jn.job_id = j.id
		  )`

	rows, err := db.connection.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching unnotified jobs for user %d: %w", userID, err)
	}
	defer rows.Close()

	var jobs []model.JobWithFilters
	for rows.Next() {
		var j model.JobWithFilters
		var filtersJSON sql.NullString
		if err := rows.Scan(&j.JobID, &j.Title, &j.Location, &j.Company, &j.JobLink, &filtersJSON); err != nil {
			return nil, fmt.Errorf("error scanning job with filters: %w", err)
		}
		if filtersJSON.Valid {
			if err := json.Unmarshal([]byte(filtersJSON.String), &j.Filters); err != nil {
				return nil, fmt.Errorf("error unmarshalling filters: %w", err)
			}
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (db *NotificationRepository) BulkInsertPendingNotifications(userID int, jobIDs []int) error {
	if len(jobIDs) == 0 {
		return nil
	}

	query := `INSERT INTO job_notifications (user_id, job_id, status) SELECT $1, unnest($2::int[]), 'PENDING' ON CONFLICT (user_id, job_id) DO NOTHING`

	_, err := db.connection.Exec(query, userID, pq.Array(jobIDs))
	if err != nil {
		return fmt.Errorf("error bulk inserting pending notifications for user %d: %w", userID, err)
	}
	return nil
}

func (db *NotificationRepository) GetUserIDsWithPendingNotifications() ([]int, error) {
	query := `SELECT DISTINCT user_id FROM job_notifications WHERE status = 'PENDING'`

	rows, err := db.connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching users with pending notifications: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}

func (db *NotificationRepository) GetPendingJobsForUser(userID int) ([]model.NotificationWithJob, error) {
	query := `
		SELECT jn.id, jn.job_id, jn.user_id, jn.notified_at,
			   j.title, j.company, j.location, j.job_link
		FROM job_notifications jn
		INNER JOIN jobs j ON jn.job_id = j.id
		WHERE jn.user_id = $1 AND jn.status = 'PENDING'
		ORDER BY j.company, j.title`

	rows, err := db.connection.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching pending jobs for user %d: %w", userID, err)
	}
	defer rows.Close()

	var notifications []model.NotificationWithJob
	for rows.Next() {
		var n model.NotificationWithJob
		if err := rows.Scan(&n.ID, &n.JobID, &n.UserID, &n.NotifiedAt,
			&n.JobTitle, &n.JobCompany, &n.JobLocation, &n.JobLink); err != nil {
			return nil, fmt.Errorf("error scanning pending notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (db *NotificationRepository) BulkUpdateNotificationStatus(userID int, jobIDs []int, status string) error {
	if len(jobIDs) == 0 {
		return nil
	}

	query := `UPDATE job_notifications SET status = $1, notified_at = NOW() WHERE user_id = $2 AND job_id = ANY($3) AND status = 'PENDING'`

	_, err := db.connection.Exec(query, status, userID, pq.Array(jobIDs))
	if err != nil {
		return fmt.Errorf("error bulk updating notification status for user %d: %w", userID, err)
	}
	return nil
}

func (db *NotificationRepository) InsertNotificationWithAnalysis(jobId int, userId int, curriculumId int, analysisResult []byte) error {
	query := `INSERT INTO job_notifications (user_id, job_id, curriculum_id, analysis_result, status) VALUES ($1, $2, $3, $4, 'SENT') ON CONFLICT (user_id, job_id) DO UPDATE SET analysis_result = $4, curriculum_id = $3, notified_at = NOW()`
	_, err := db.connection.Exec(query, userId, jobId, curriculumId, analysisResult)
	if err != nil {
		return fmt.Errorf("error inserting notification with analysis: %w", err)
	}
	return nil
}

func (db *NotificationRepository) GetAnalysisHistory(userId int, jobId int) ([]byte, *int, error) {
	query := `SELECT analysis_result, curriculum_id FROM job_notifications WHERE user_id = $1 AND job_id = $2 AND analysis_result IS NOT NULL ORDER BY notified_at DESC LIMIT 1`
	var result []byte
	var curriculumID sql.NullInt64
	err := db.connection.QueryRow(query, userId, jobId).Scan(&result, &curriculumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("error fetching analysis history: %w", err)
	}
	var cvID *int
	if curriculumID.Valid {
		id := int(curriculumID.Int64)
		cvID = &id
	}
	return result, cvID, nil
}
