package tasks

import "web-scrapper/model"

const (
	TypeScrapSite      = "scrape:site"
	TypeProcessResults = "process:results"
	TypeNotifyUser     = "notify:user"
	TypeAnalyzeUserJob = "analyze:resume"
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