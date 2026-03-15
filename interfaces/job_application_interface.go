package interfaces

import (
	"web-scrapper/model"
)

type JobApplicationRepositoryInterface interface {
	Create(userID, jobID int) (model.JobApplication, error)
	Update(id, userID int, req model.UpdateApplicationRequest) (model.JobApplication, error)
	Delete(id, userID int) error
	GetAllByUser(userID int) ([]model.JobApplicationWithJob, error)
	ExistsByUserAndJob(userID, jobID int) (bool, error)
	JobExists(jobID int) (bool, error)
}
