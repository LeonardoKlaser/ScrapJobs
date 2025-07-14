package tasks

import "web-scrapper/model"

const (
	TypeScrapSite      = "scrape:site"
	TypeProcessResults = "process:results"
)

type ScrapeSitePayload struct {
	SiteID             int
	SiteScrapingConfig model.SiteScrapingConfig
}

type ProcessResultsPayload struct {
	SiteID int
	Jobs []*model.Job
}