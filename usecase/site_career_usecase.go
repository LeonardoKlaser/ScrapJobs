package usecase

import (
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/scrapper"
	"context"
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

func (repo *SiteCareerUsecase) SandboxScrape(ctx context.Context,config model.SiteScrapingConfig) ([]*model.Job, error){
	scrapInterface, err := scrapper.NewScraperFactory(config)
    if err != nil {
        return nil, err
    }

	jobs, err := scrapInterface.Scrape(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro durante o processo de scraping: %w ", err)
	}

	return jobs, nil
}