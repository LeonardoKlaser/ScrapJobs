package interfaces

import(
	"web-scrapper/model"
)

type JobRepositoryInterface interface{
	CreateJob(job model.Job) (int, error)
	FindJobByRequisitionID(requisition_ID int) (bool, error)
	FindJobsByRequisitionIDs(requisition_IDs []int64) (map[int64]bool, error)
	UpdateLastSeen(requisition_ID int64) (int, error)
}