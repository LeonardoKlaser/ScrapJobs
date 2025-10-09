package usecase

import (
	"web-scrapper/repository"
)

type RequestedSiteUsecase struct {
	repo *repository.RequestedSiteRepository
}

func NewRequestedSiteUsecase(repo *repository.RequestedSiteRepository) *RequestedSiteUsecase {
	return &RequestedSiteUsecase{
		repo: repo,
	}
}

func (uc *RequestedSiteUsecase) Create(userID int, url string) error {
	return uc.repo.Create(userID, url)
}