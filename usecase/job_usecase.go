package usecase

import (
	"context"
	"log"
	"web-scrapper/model"
	"web-scrapper/interfaces"
	"web-scrapper/scrapper"
)

type JobUseCase struct{
	Repository interfaces.JobRepositoryInterface
}

func NewJobUseCase(jobRepo interfaces.JobRepositoryInterface) *JobUseCase{
	return &JobUseCase{
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


func (uc *JobUseCase) ScrapeAndStoreJobs(ctx context.Context, selectors model.SiteScrapingConfig) ([]*model.Job, error) {
    scrapInterface, err := scrapper.NewScraperFactory(selectors)
    if err != nil {
        return nil, err
    }

	jobs, err := scrapInterface.Scrape(ctx, selectors)
	
    var newJobsToDatabase []*model.Job
	ids := takeIDs(jobs)
	exist, err := uc.Repository.FindJobsByRequisitionIDs(ids)
	if err != nil{
		return nil, err
	}
    for _, job := range jobs {
        if _ , ok := exist[job.Requisition_ID]; !ok {
			jobToInsert := model.Job{
				Title: job.Title,
				Location: job.Location,
				Company: job.Company,
				Job_link: job.Job_link,
				Requisition_ID: job.Requisition_ID,
			}
            ID, err := uc.Repository.CreateJob(jobToInsert)
			if err != nil {
				log.Printf("Error to create job %v : %v", job, err)
			}
			job.ID = ID
			newJobsToDatabase = append(newJobsToDatabase, job)
        }else{
			jobID, err := uc.Repository.UpdateLastSeen(job.Requisition_ID)
			if err != nil {
				log.Printf("Error to update last seen job %v : %v", job, err)
			}
			job.ID = jobID
			newJobsToDatabase = append(newJobsToDatabase, job)
		}
    }
    return newJobsToDatabase, nil
}

func takeIDs(jobs []*model.Job) []int64{
	var ids []int64
	for _, job := range(jobs){
		ids = append(ids, job.Requisition_ID)
	}
	return ids
}