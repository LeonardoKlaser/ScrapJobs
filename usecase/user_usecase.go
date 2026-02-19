package usecase

import (
	"errors"
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/model"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repository interfaces.UserRepositoryInterface
}

func NewUserUsercase(repo interfaces.UserRepositoryInterface) *UserUsecase {
	return &UserUsecase{
		repository: repo,
	}
}

func (usr *UserUsecase) CreateUser(user model.User) (model.User, error) {
	exist, err := usr.repository.GetUserByEmail(user.Email)
	if err != nil {
		return model.User{}, err
	}

	if exist.Email != "" {
		return model.User{}, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("erro ao criptografar a senha: %w", err)
	}

	user.Password = string(hashedPassword)

	created, err := usr.repository.CreateUser(user)
	if err != nil {
		return model.User{}, err
	}
	return created, nil
}

func (usr *UserUsecase) CreateUserWithHashedPassword(user model.User) (model.User, error) {
	exist, err := usr.repository.GetUserByEmail(user.Email)
	if err != nil {
		return model.User{}, err
	}

	if exist.Email != "" {
		return model.User{}, errors.New("user already exists")
	}

	created, err := usr.repository.CreateUser(user)
	if err != nil {
		return model.User{}, err
	}
	return created, nil
}

func (usr *UserUsecase) GetUserByEmail(userEmail string) (model.User, error) {
	res, err := usr.repository.GetUserByEmail(userEmail)
	if err != nil {
		return model.User{}, err
	}
	return res, nil
}

func (usr *UserUsecase) GetUserById(Id int) (model.User, error) {
	res, err := usr.repository.GetUserById(Id)
	if err != nil {
		return model.User{}, err
	}
	return res, nil
}

func (usr *UserUsecase) UpdateUserProfile(userId int, name string, cellphone *string, tax *string) error {
	return usr.repository.UpdateUserProfile(userId, name, cellphone, tax)
}

func (usr *UserUsecase) ChangePassword(userId int, currentHash, oldPassword, newPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword))
	if err != nil {
		return errors.New("senha atual incorreta")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erro ao criptografar a senha: %w", err)
	}

	return usr.repository.UpdateUserPassword(userId, string(hashedPassword))
}
