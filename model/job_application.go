package model

import "time"

type JobApplication struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	JobID          int       `json:"job_id"`
	Status         string    `json:"status"`
	InterviewRound *int      `json:"interview_round"`
	Notes          *string   `json:"notes"`
	AppliedAt      time.Time `json:"applied_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type JobApplicationWithJob struct {
	JobApplication
	Job JobApplicationJob `json:"job"`
}

type JobApplicationJob struct {
	Title    string `json:"title"`
	Company  string `json:"company"`
	Location string `json:"location"`
	JobLink  string `json:"job_link"`
}

type CreateApplicationRequest struct {
	JobID int `json:"job_id" binding:"required"`
}

type UpdateApplicationRequest struct {
	Status         *string `json:"status"`
	InterviewRound *int    `json:"interview_round"`
	Notes          *string `json:"notes"`
}

type ApplicationsResponse struct {
	Applications []JobApplicationWithJob `json:"applications"`
}
