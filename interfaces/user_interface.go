package interfaces

import (
	"time"
	"web-scrapper/model"
)

type UserRepositoryInterface interface {
	CreateUser(user model.User) (model.User, error)
	GetUserByEmail(userEmail string) (model.User, error)
	GetUserById(Id int) (model.User, error)
	UpdateUserProfile(userId int, name string, cellphone *string, tax *string) error
	UpdateUserPassword(userId int, hashedPassword string) error
	CheckUserExists(email string, tax string) (bool, bool, error)
	GetUserBasicInfo(userID int) (string, string, error)
	SoftDeleteUser(userId int) error
	UpdateExpiresAt(userId int, expiresAt time.Time) error
	UpdateWeekdaysOnly(userID int, value bool) error
}
