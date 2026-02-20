package mocks

import (
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

func (m *MockUserRepository) UpdateUserProfile(userId int, name string, cellphone *string, tax *string) error {
	args := m.Called(userId, name, cellphone, tax)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserPassword(userId int, hashedPassword string) error {
	args := m.Called(userId, hashedPassword)
	return args.Error(0)
}
