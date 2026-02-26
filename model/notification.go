package model

import "time"

// NotificationWithJob representa uma notificação com dados da vaga associada
type NotificationWithJob struct {
	ID          int       `json:"id"`
	JobID       int       `json:"job_id"`
	UserID      int       `json:"user_id"`
	NotifiedAt  time.Time `json:"notified_at"`
	JobTitle    string    `json:"job_title"`
	JobCompany  string    `json:"job_company"`
	JobLocation string    `json:"job_location"`
	JobLink     string    `json:"job_link"`
}

// JobWithFilters representa uma vaga com os filtros do usuário associados
type JobWithFilters struct {
	JobID    int      `json:"job_id"`
	Title    string   `json:"title"`
	Location string   `json:"location"`
	Company  string   `json:"company"`
	JobLink  string   `json:"job_link"`
	Filters  []string `json:"-"`
}
