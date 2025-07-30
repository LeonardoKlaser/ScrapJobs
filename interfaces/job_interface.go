package interfaces

import(
	"web-scrapper/model"
)

type JobRepositoryInterface interface{
	CreateJob(job model.Job) (int, error)
	FindJobByRequisitionID(requisition_ID int) (bool, error)
	FindJobsByRequisitionIDs(requisition_IDs []int) (map[int]bool, error)
	UpdateLastSeen(requisition_ID int) (int, error)
}