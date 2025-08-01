package usecase

import (
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/scrapper"
)

type SiteCareerUsecase struct{
	repo interfaces.SiteCareerRepositoryInterface
}

func NewSiteCareerUsecase(repo interfaces.SiteCareerRepositoryInterface) *SiteCareerUsecase {
	return &SiteCareerUsecase{
		repo: repo,
	}
}

func (repo *SiteCareerUsecase) InsertNewSiteCareer(site model.SiteScrapingConfig) (model.SiteScrapingConfig, error){
	res, err := repo.repo.InsertNewSiteCareer(site)
	if err != nil {
		return model.SiteScrapingConfig{}, err
	}

	return res, nil
}

func (repo *SiteCareerUsecase) SandboxScrape(config model.SiteScrapingConfig) ([]*model.Job, error){
	scraper := scrapper.NewJobScraper()
	
	jobs, err := scraper.ScrapeJobs(config)
	if err != nil {
		return nil, fmt.Errorf("erro durante o processo de scraping: %w ", err)
	}

	return jobs, nil
}