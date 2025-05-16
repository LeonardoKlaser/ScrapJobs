package usecase

import (
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/scrapper"
)

type JobUseCase struct{
	Repository repository.JobRepository
	scrapper scrapper.JobScrapper
}

func NewJobUseCase(jobRepo repository.JobRepository) JobUseCase{
	return JobUseCase{
		Repository: jobRepo,
	}
}

func (job JobUseCase) CreateJob(jobData model.Job) (int, error){
	jobID , err := job.Repository.CreateJob(jobData);
	if(err != nil){
		return jobID, err
	}

	return jobID, nil
}

func (job JobUseCase) FindJobByRequisitionID(requisition_ID int) (bool, error){
	hasJob, err := job.Repository.FindJobByRequisitionID(requisition_ID);
	if(err != nil){
		return false, err
	}

	return hasJob, nil
}


func (uc *JobUseCase) ScrapeAndStoreJobs() ([]*model.Job, error) {
    jobs, err := scrapper.NewJobScraper().ScrapeJobs()
    if err != nil {
        return nil, err
    }
    var newJobsToDatabase []*model.Job
    for _, job := range jobs {
        exist, err := uc.Repository.FindJobByRequisitionID(job.Requisition_ID)
        if err != nil {
            return nil, err
        }
        if !exist {
            newJobsToDatabase = append(newJobsToDatabase, job)
			jobToInsert := model.Job{
				Title: job.Title,
				Location: job.Location,
				Company: job.Company,
				Job_link: job.Job_link,
				Requisition_ID: job.Requisition_ID,
			}
            uc.Repository.CreateJob(jobToInsert)
        }
    }
    return newJobsToDatabase, nil
}
