package usecase

import (
	"context"
	"time"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"
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
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    scrapInterface, err := scrapper.NewScraperFactory(selectors)
    if err != nil {
        return nil, err
    }

	jobs, err := scrapInterface.Scrape(ctx, selectors)
	if err != nil {
		return []*model.Job{}, err
	}

    var newJobsToDatabase []*model.Job
	ids := takeIDs(jobs)
	exist, err := uc.Repository.FindJobsByRequisitionIDs(ids)
	if err != nil{
		return nil, err
	}
    for _, job := range jobs {
        if _ , ok := exist[job.RequisitionID]; !ok {
			jobToInsert := model.Job{
				Title: job.Title,
				Location: job.Location,
				Company: selectors.SiteName,
				JobLink: job.JobLink,
				RequisitionID: job.RequisitionID,
			}
            ID, err := uc.Repository.CreateJob(jobToInsert)
			if err != nil {
				logging.Logger.Error().Err(err).Str("job_title", job.Title).Msg("Failed to create job")
			}
			job.ID = ID
			newJobsToDatabase = append(newJobsToDatabase, job)
        }else{
			jobID, err := uc.Repository.UpdateLastSeen(job.RequisitionID)
			if err != nil {
				logging.Logger.Error().Err(err).Int64("requisition_id", job.RequisitionID).Msg("Failed to update last seen")
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
		ids = append(ids, job.RequisitionID)
	}
	return ids
}
