package usecase

import (
	"fmt"
	"web-scrapper/interfaces"
)

type UserSiteUsecase struct {
	rep      interfaces.UserSiteRepositoryInterface
	planRepo interfaces.PlanRepositoryInterface
}

func NewUserSiteUsecase(rep interfaces.UserSiteRepositoryInterface, planRepo interfaces.PlanRepositoryInterface) *UserSiteUsecase {
	return &UserSiteUsecase{
		rep:      rep,
		planRepo: planRepo,
	}
}

func (usu *UserSiteUsecase) InsertUserSite(userId int, siteId int, filters []string) error {
	plan, err := usu.planRepo.GetPlanByUserID(userId)
	if err != nil {
		return fmt.Errorf("erro ao buscar plano do usuário: %w", err)
	}

	if plan == nil {
		return fmt.Errorf("nenhum plano associado ao usuário. Assine um plano para monitorar sites")
	}

	count, err := usu.rep.GetUserSiteCount(userId)
	if err != nil {
		return fmt.Errorf("erro ao contar sites do usuário: %w", err)
	}

	if count >= plan.MaxSites {
		return fmt.Errorf("limite de sites atingido (%d/%d). Faça upgrade do seu plano para monitorar mais sites", count, plan.MaxSites)
	}

	return usu.rep.InsertNewUserSite(userId, siteId, filters)
}

func (usu *UserSiteUsecase) DeleteUserSite(userId int, siteId string) error {
	return usu.rep.DeleteUserSite(userId, siteId)
}

// UpdateUserSiteFilters atualiza os filtros (palavras-chave) de monitoramento de um site
func (usu *UserSiteUsecase) UpdateUserSiteFilters(userId int, siteId int, filters []string) error {
	return usu.rep.UpdateUserSiteFilters(userId, siteId, filters)
}
