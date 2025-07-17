package usecase

import (
	"web-scrapper/model"
	"web-scrapper/repository"
)

type UserUsecase struct {
	repository repository.UserRepository
}

func NewUserUsercase (repo repository.UserRepository) UserUsecase{
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