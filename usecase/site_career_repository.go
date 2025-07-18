package usecase

import (
	"web-scrapper/model"
	"web-scrapper/repository"
)

type SiteCareerUsecase struct{
	repo *repository.SiteCareerRepository
}

func NewSiteCareerUsecase(repo *repository.SiteCareerRepository) *SiteCareerUsecase {
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