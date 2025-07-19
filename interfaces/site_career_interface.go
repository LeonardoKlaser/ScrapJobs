package interfaces

import "web-scrapper/model"

type SiteCareerRepositoryInterface interface {
	InsertNewSiteCareer(site model.SiteScrapingConfig) (model.SiteScrapingConfig, error)
	GetAllSites() ([]model.SiteScrapingConfig, error)
}