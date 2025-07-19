package usecase

import (
	"web-scrapper/model"
	"web-scrapper/interfaces"
)

type UserUsecase struct {
	repository interfaces.UserRepositoryInterface
}

func NewUserUsercase (repo interfaces.UserRepositoryInterface) UserUsecase{
	return UserUsecase{
		repository: repo,
	}
}

func (usr *UserUsecase) CreateUser(user model.User) (model.User, error){
	res, err := usr.repository.CreateUser(user)
	if err != nil{
		return model.User{}, err
	}
	return res, nil
}

func (usr *UserUsecase) GetUserByEmail(userEmail string) (model.User, error){
	res, err := usr.repository.GetUserByEmail(userEmail)
	if err != nil{
		return model.User{}, err
	}
	return res, nil
}

func (usr *UserUsecase) GetUserById(Id int) (model.User, error){
	res, err := usr.repository.GetUserById(Id)
	if err != nil{
		return model.User{}, err
	}
	return res, nil
}