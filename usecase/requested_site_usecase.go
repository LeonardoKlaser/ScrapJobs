package usecase

import (
	"web-scrapper/interfaces"
)

type RequestedSiteUsecase struct {
	repo interfaces.RequestedSiteRepositoryInterface
}

func NewRequestedSiteUsecase(repo interfaces.RequestedSiteRepositoryInterface) *RequestedSiteUsecase {
	return &RequestedSiteUsecase{
		repo: repo,
	}
}

func (uc *RequestedSiteUsecase) Create(userID int, url string) error {
	return uc.repo.Create(userID, url)
}