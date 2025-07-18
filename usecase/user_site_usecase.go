package usecase

import(
	"web-scrapper/repository"
)

type UserSiteUsecase struct{
	rep *repository.UserSiteRepository
}

func NewUserSiteUsecase(rep *repository.UserSiteRepository) *UserSiteUsecase{
	return &UserSiteUsecase{
		rep: rep,
	}
}


func (rep *UserSiteUsecase) InsertUserSite(userId int, siteId int, filters []string) error{
	err := rep.rep.InsertNewUserSite(userId, siteId, filters)
	if err != nil {
		return err
	}

	return nil
}