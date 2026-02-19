package model

type Job struct {
	ID            int    `json:"id" db:"id"`
	Title         string `json:"title" db:"title"`
	Location      string `json:"location" db:"location"`
	Company       string `json:"company" db:"company"`
	JobLink       string `json:"job_link" db:"job_link"`
	RequisitionID int64  `json:"job_id" db:"requisition_id"`
	Description   string `json:"description" db:"description"`
}
