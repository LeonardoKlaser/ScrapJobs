package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"web-scrapper/infra/s3"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/scrapper"
)

type SiteCareerUsecase struct{
	repo interfaces.SiteCareerRepositoryInterface
	s3Uploader s3.UploaderInterface
}

func NewSiteCareerUsecase(repo interfaces.SiteCareerRepositoryInterface, uploader s3.UploaderInterface) *SiteCareerUsecase {
	return &SiteCareerUsecase{
		repo: repo,
		s3Uploader: uploader,

	}
}

func (repo *SiteCareerUsecase) InsertNewSiteCareer(ctx context.Context ,site model.SiteScrapingConfig, file *multipart.FileHeader) (model.SiteScrapingConfig, error){

	if file != nil {
		logoURL, err := repo.s3Uploader.UploadFile(ctx, file)
		if err != nil {
			return model.SiteScrapingConfig{}, fmt.Errorf("falha no upload do logo: %w", err)
		}
		site.LogoURL = logoURL
	}

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

func (repo *SiteCareerUsecase) GetAllSites() ([]model.SiteScrapingConfig, error){
	sites, err := repo.repo.GetAllSites()
	if err != nil {
		return nil, err
	}

	return sites, nil;
}