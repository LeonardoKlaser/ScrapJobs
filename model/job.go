package model

type Job struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Location string `json:"location"`
	Company string `json:"company"`
	Job_link string `json:"job_link"`
	Requisition_ID int `json:"job_id"`
}