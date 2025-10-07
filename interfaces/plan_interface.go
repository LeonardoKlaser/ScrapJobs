package interfaces

import "web-scrapper/model"

type PlanRepositoryInterface interface {
	GetAllPlans() ([]model.Plan, error)
	GetPlanByUserID(userID int) (*model.Plan, error)
}