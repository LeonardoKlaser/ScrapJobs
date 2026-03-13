package mocks

import (
	"time"
	"web-scrapper/model"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user model.User) (model.User, error) {
	args := m.Called(user)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(userEmail string) (model.User, error) {
	args := m.Called(userEmail)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserById(Id int) (model.User, error) {
	args := m.Called(Id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserMeData(userID int) (model.UserMeData, error) {
	args := m.Called(userID)
	return args.Get(0).(model.UserMeData), args.Error(1)
}

func (m *MockUserRepository) UpdateUserProfile(userId int, name string, cellphone *string, tax *string) error {
	args := m.Called(userId, name, cellphone, tax)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserPassword(userId int, hashedPassword string) error {
	args := m.Called(userId, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) CheckUserExists(email string, tax string) (bool, bool, error) {
	args := m.Called(email, tax)
	return args.Bool(0), args.Bool(1), args.Error(2)
}

func (m *MockUserRepository) GetUserBasicInfo(userID int) (string, string, error) {
	args := m.Called(userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserRepository) SoftDeleteUser(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateExpiresAt(userId int, expiresAt time.Time) error {
	args := m.Called(userId, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateWeekdaysOnly(userID int, value bool) error {
	args := m.Called(userID, value)
	return args.Error(0)
}
