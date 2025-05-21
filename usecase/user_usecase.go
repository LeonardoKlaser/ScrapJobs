package usecase

import (
	"context"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/scrapper"
)

type JobUseCase struct{
	Repository repository.JobRepository
	scrapper scrapper.JobScrapper
	MailSender *ses.SESMailSender
}

func NewJobUseCase(jobRepo repository.JobRepository, mailSender *ses.SESMailSender) JobUseCase{
	return JobUseCase{
		Repository: jobRepo,
		MailSender: mailSender,
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


func (uc *JobUseCase) ScrapeAndStoreJobs(ctx context.Context) ([]*model.Job, error) {
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
			if uc.MailSender != nil{
				subject := "Nova Vaga Encontrada: " + job.Title
				body := "Uma nova vaga foi encontrada: " + job.Title + "\n" +
					"Link para saber mais sobre a vaga: " + "https://jobs.sap.com" + job.Job_link + "\n" 
				to := "leobkklaser@gmail.com"

				go func() {
					err := uc.MailSender.SendEmail(ctx, to, subject, body)
					if err != nil {
						println("Erro ao enviar e-mail:", err)
						return
					}
				}()
			}
        }
    }
    return newJobsToDatabase, nil
}
