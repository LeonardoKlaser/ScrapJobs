package interfaces

import (
	"web-scrapper/model"
)

type JobRepositoryInterface interface {
	CreateJob(job model.Job) (int, error)
	FindJobByRequisitionID(requisition_ID string) (bool, error)
	FindJobsByRequisitionIDs(requisition_IDs []string) (map[string]bool, error)
	UpdateLastSeen(requisition_ID string) (int, error)
	DeleteOldJobs() error
	GetJobByID(jobID int) (*model.Job, error)
}
