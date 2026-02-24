package tasks

import "web-scrapper/model"

const (
	TypeScrapSite           = "scrape:site"
	TypeProcessResults      = "process:results"
	TypeNotifyUser          = "notify:user"
	TypeAnalyzeUserJob      = "analyze:resume"
	TypeCompleteRegistration = "payment:complete_registration"
	TypeNotifyNewJobs       = "notify:new_jobs"
)

type ScrapeSitePayload struct {
	SiteID             int
	SiteScrapingConfig model.SiteScrapingConfig
}

type ProcessResultsPayload struct {
	SiteID int
	Jobs []*model.Job
}

type NotifyUserPayload struct {
	User model.UserSiteCurriculum
	Job *model.Job
	Analysis model.ResumeAnalysis
}

type AnalyzeUserJobPayload struct {
	User model.UserSiteCurriculum
	Job *model.Job
}

type NotifyNewJobsPayload struct {
	User model.UserSiteCurriculum
	Jobs []*model.Job
}

type CompleteRegistrationPayload struct {
	PendingRegistrationID string `json:"pending_registration_id"`
	CustomerEmail         string `json:"customer_email"`
}