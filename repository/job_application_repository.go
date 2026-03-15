package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"web-scrapper/model"

	"github.com/lib/pq"
)

var (
	ErrApplicationNotFound = errors.New("candidatura não encontrada")
	ErrApplicationExists   = errors.New("candidatura já existe para esta vaga")
)

type JobApplicationRepository struct {
	connection *sql.DB
}

func NewJobApplicationRepository(db *sql.DB) *JobApplicationRepository {
	return &JobApplicationRepository{connection: db}
}

func (r *JobApplicationRepository) Create(userID, jobID int) (model.JobApplication, error) {
	var app model.JobApplication
	query := `
		INSERT INTO job_applications (user_id, job_id)
		VALUES ($1, $2)
		RETURNING id, user_id, job_id, status, interview_round, notes, applied_at, updated_at`

	err := r.connection.QueryRow(query, userID, jobID).Scan(
		&app.ID, &app.UserID, &app.JobID, &app.Status,
		&app.InterviewRound, &app.Notes, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return app, ErrApplicationExists
		}
		return app, fmt.Errorf("erro ao criar candidatura: %w", err)
	}
	return app, nil
}

func (r *JobApplicationRepository) Update(id, userID int, req model.UpdateApplicationRequest) (model.JobApplication, error) {
	var app model.JobApplication

	// interview_round update logic:
	// - When status is provided ($1 IS NOT NULL): set interview_round to $2
	//   (will be the round number for "interview" status, or NULL for other statuses
	//    since the controller clears InterviewRound for non-interview statuses)
	// - When status is NOT provided (notes-only update): preserve existing interview_round
	query := `
		UPDATE job_applications
		SET status = COALESCE($1, status),
		    interview_round = CASE WHEN $1 IS NOT NULL THEN $2 ELSE interview_round END,
		    notes = COALESCE($3, notes),
		    updated_at = NOW()
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, job_id, status, interview_round, notes, applied_at, updated_at`

	err := r.connection.QueryRow(query, req.Status, req.InterviewRound, req.Notes, id, userID).Scan(
		&app.ID, &app.UserID, &app.JobID, &app.Status,
		&app.InterviewRound, &app.Notes, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return app, ErrApplicationNotFound
		}
		return app, fmt.Errorf("erro ao atualizar candidatura: %w", err)
	}
	return app, nil
}

func (r *JobApplicationRepository) Delete(id, userID int) error {
	result, err := r.connection.Exec(
		`DELETE FROM job_applications WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("erro ao remover candidatura: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrApplicationNotFound
	}
	return nil
}

func (r *JobApplicationRepository) GetAllByUser(userID int) ([]model.JobApplicationWithJob, error) {
	query := `
		SELECT ja.id, ja.user_id, ja.job_id, ja.status, ja.interview_round, ja.notes, ja.applied_at, ja.updated_at,
		       j.title, j.company, j.location, j.job_link
		FROM job_applications ja
		JOIN jobs j ON j.id = ja.job_id
		WHERE ja.user_id = $1
		ORDER BY ja.updated_at DESC`

	rows, err := r.connection.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar candidaturas: %w", err)
	}
	defer rows.Close()

	var apps []model.JobApplicationWithJob
	for rows.Next() {
		var a model.JobApplicationWithJob
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.JobID, &a.Status, &a.InterviewRound, &a.Notes, &a.AppliedAt, &a.UpdatedAt,
			&a.Job.Title, &a.Job.Company, &a.Job.Location, &a.Job.JobLink,
		); err != nil {
			return nil, fmt.Errorf("erro ao ler candidatura: %w", err)
		}
		apps = append(apps, a)
	}

	if apps == nil {
		apps = []model.JobApplicationWithJob{}
	}

	return apps, rows.Err()
}

func (r *JobApplicationRepository) ExistsByUserAndJob(userID, jobID int) (bool, error) {
	var exists bool
	err := r.connection.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM job_applications WHERE user_id = $1 AND job_id = $2)`,
		userID, jobID,
	).Scan(&exists)
	return exists, err
}

func (r *JobApplicationRepository) JobExists(jobID int) (bool, error) {
	var exists bool
	err := r.connection.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM jobs WHERE id = $1)`,
		jobID,
	).Scan(&exists)
	return exists, err
}
