package interfaces

import "web-scrapper/model"

type UserRepositoryInterface interface {
	CreateUser(user model.User) (model.User, error)
	GetUserByEmail(userEmail string) (model.User, error)
	GetUserById(Id int) (model.User, error)
}
