package usecase

import (
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_CreateUser(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)

	userUsecase := NewUserUsercase(mockRepo)

	t.Run("should create user successfully", func(t *testing.T) {
		newUser := model.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("GetUserByEmail", "test@example.com").Return(model.User{}, nil).Once()
		mockRepo.On("CreateUser", mock.AnythingOfType("model.User")).Return(model.User{
			Id:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}, nil).Once()

		created, err := userUsecase.CreateUser(newUser)

		assert.NoError(t, err)
		assert.Equal(t, 1, created.Id)
		assert.Equal(t, "Test User", created.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should fail when user already exists", func(t *testing.T) {
		existingUser := model.User{
			Name:     "Existing User",
			Email:    "exists@example.com",
			Password: "password123",
		}

		mockRepo.On("GetUserByEmail", "exists@example.com").Return(model.User{
			Id:    2,
			Email: "exists@example.com",
		}, nil).Once()

		_, err := userUsecase.CreateUser(existingUser)

		assert.Error(t, err)
		assert.Equal(t, "user already exists", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_ChangePassword(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)

	userUsecase := NewUserUsercase(mockRepo)

	currentHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)

	t.Run("should change password successfully", func(t *testing.T) {
		mockRepo.On("UpdateUserPassword", 1, mock.AnythingOfType("string")).Return(nil).Once()

		err := userUsecase.ChangePassword(1, string(currentHash), "oldpassword", "newpassword123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should fail with wrong current password", func(t *testing.T) {
		err := userUsecase.ChangePassword(1, string(currentHash), "wrongpassword", "newpassword123")

		assert.Error(t, err)
		assert.Equal(t, "senha atual incorreta", err.Error())
	})
}

func TestUserUsecase_UpdateProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)

	userUsecase := NewUserUsercase(mockRepo)

	t.Run("should update profile successfully", func(t *testing.T) {
		phone := "11999999999"
		tax := "12345678901"
		mockRepo.On("UpdateUserProfile", 1, "New Name", &phone, &tax).Return(nil).Once()

		err := userUsecase.UpdateUserProfile(1, "New Name", &phone, &tax)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should update profile with nil optionals", func(t *testing.T) {
		mockRepo.On("UpdateUserProfile", 1, "Name Only", (*string)(nil), (*string)(nil)).Return(nil).Once()

		err := userUsecase.UpdateUserProfile(1, "Name Only", nil, nil)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_GetUserById(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)

	userUsecase := NewUserUsercase(mockRepo)

	t.Run("should get user by id", func(t *testing.T) {
		expected := model.User{Id: 1, Name: "Test", Email: "test@test.com"}
		mockRepo.On("GetUserById", 1).Return(expected, nil).Once()

		user, err := userUsecase.GetUserById(1)

		assert.NoError(t, err)
		assert.Equal(t, expected.Id, user.Id)
		assert.Equal(t, expected.Name, user.Name)
		mockRepo.AssertExpectations(t)
	})
}
