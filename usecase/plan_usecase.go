package usecase

import (
	"web-scrapper/interfaces"
	"web-scrapper/model"
)

type PlanUsecase struct {
	repository interfaces.PlanRepositoryInterface
}

func NewPlanUsecase(repo interfaces.PlanRepositoryInterface) *PlanUsecase {
	return &PlanUsecase{
		repository: repo,
	}
}

func (uc *PlanUsecase) GetAllPlans() ([]model.Plan, error) {
	return uc.repository.GetAllPlans()
}

func (uc *PlanUsecase) GetPlanByUserID(userID int) (*model.Plan, error) {
	return uc.repository.GetPlanByUserID(userID)
}

func (uc *PlanUsecase) GetPlanByID(planId int) (*model.Plan, error) {
	return uc.repository.GetPlanByID(planId)
}