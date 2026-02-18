package model

// NotificationWithJob representa uma notificação com dados da vaga associada
type NotificationWithJob struct {
	ID          int    `json:"id"`
	JobID       int    `json:"job_id"`
	UserID      int    `json:"user_id"`
	NotifiedAt  string `json:"notified_at"`
	JobTitle    string `json:"job_title"`
	JobCompany  string `json:"job_company"`
	JobLocation string `json:"job_location"`
	JobLink     string `json:"job_link"`
}
