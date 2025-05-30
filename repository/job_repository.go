package repository

import (
	"database/sql"
	"web-scrapper/model"
)
type JobRepository struct {
	connection *sql.DB
}

func NewJobRepository(db *sql.DB) JobRepository {
	return JobRepository{
		connection: db,
	}
}

func (usr *JobRepository) CreateJob(job model.Job) (int, error) {
	query := `INSERT INTO jobs (title, location, company, job_link, requisition_ID) VALUES ($1, $2, $3, $4, $5)`
	queryPrepare, err := usr.connection.Prepare(query)

	if(err != nil){
		return 0, err
	}

	err = queryPrepare.QueryRow(job.Title, job.Location, job.Company, job.Job_link, job.Requisition_ID).Scan(&job.ID)
	if(err != nil){
		return 0, err	
	}

	queryPrepare.Close()
	return job.ID, nil
}

func (usr *JobRepository) FindJobByRequisitionID(requisition_ID int) (bool, error){
	query := `SELECT COUNT(*) FROM jobs WHERE requisition_ID = $1`
	queryPrepare, err := usr.connection.Prepare(query)
	if(err != nil){
		return false, err
	}

	var count int
	err = queryPrepare.QueryRow(requisition_ID).Scan(&count)
	if(err != nil){
		return false, err
	}

	queryPrepare.Close()
	return count > 0, nil
}