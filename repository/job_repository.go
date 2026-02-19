package repository

import (
	"database/sql"
	"fmt"
	"log"
	"web-scrapper/model"

	"github.com/lib/pq"
)

type JobRepository struct {
	connection *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{
		connection: db,
	}
}

func (usr *JobRepository) CreateJob(job model.Job) (int, error) {
	query := `INSERT INTO jobs (title, location, company, job_link, requisition_ID) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	queryPrepare, err := usr.connection.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer queryPrepare.Close()

	err = queryPrepare.QueryRow(job.Title, job.Location, job.Company, job.Job_link, job.Requisition_ID).Scan(&job.ID)
	if err != nil {
		return 0, err
	}

	return job.ID, nil
}

func (usr *JobRepository) FindJobByRequisitionID(requisition_ID int) (bool, error) {
	query := `SELECT COUNT(*) FROM jobs WHERE requisition_ID = $1`
	queryPrepare, err := usr.connection.Prepare(query)
	if err != nil {
		return false, err
	}
	defer queryPrepare.Close()

	var count int
	err = queryPrepare.QueryRow(requisition_ID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (usr *JobRepository) FindJobsByRequisitionIDs(requisition_IDs []int64) (map[int64]bool, error){
	query := `SELECT requisition_ID FROM jobs WHERE requisition_ID = ANY($1)`
	
	Exists := make(map[int64]bool)
    if len(requisition_IDs) == 0 {
        return Exists, nil
    }

	rows, err := usr.connection.Query(query, pq.Array(requisition_IDs) )
	if err != nil {
		return Exists, fmt.Errorf("error fetching jobs %d: %w", requisition_IDs, err)
    }
	
	defer rows.Close()

	for rows.Next(){
		var requisitionID int64
		if err := rows.Scan(&requisitionID); err != nil {
			return nil, fmt.Errorf("error scanning notified job requisition ID: %w", err)
		}
		Exists[requisitionID] = true
	}

	return Exists, rows.Err()
}

func (usr *JobRepository) UpdateLastSeen(requisition_ID int64) (int, error) {
	var id int
	query := `UPDATE jobs SET last_seen_at = CURRENT_TIMESTAMP WHERE requisition_ID = $1 RETURNING id`
	queryPrepare, err := usr.connection.Prepare(query)
	if err != nil {
		log.Printf("error to prepare query to update last seen for job id: %d - %v", requisition_ID, err)
		return id, err
	}
	defer queryPrepare.Close()

	err = queryPrepare.QueryRow(requisition_ID).Scan(&id)
	if err != nil {
		log.Printf("error to update last seen for job id: %d - %v", requisition_ID, err)
		return id, err
	}

	log.Printf("last seen updated for job id: %d", requisition_ID)
	return id, nil
}

func (usr *JobRepository) GetJobByID(jobID int) (*model.Job, error) {
	query := `SELECT id, title, location, company, job_link, requisition_id, description FROM jobs WHERE id = $1`

	var job model.Job
	err := usr.connection.QueryRow(query, jobID).Scan(
		&job.ID,
		&job.Title,
		&job.Location,
		&job.Company,
		&job.Job_link,
		&job.Requisition_ID,
		&job.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching job by ID %d: %w", jobID, err)
	}
	return &job, nil
}

func (usr *JobRepository) DeleteOldJobs() error {
	query := `DELETE FROM jobs WHERE last_seen_at < NOW() - INTERVAL '1 day'`

	result, err := usr.connection.Exec(query)
	if err != nil {
		return fmt.Errorf("error deleting old jobs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	log.Printf("deleted %d old jobs", rowsAffected)
	return nil
}