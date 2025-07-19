package usecase

import (
	"web-scrapper/model"
	"web-scrapper/interfaces"
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