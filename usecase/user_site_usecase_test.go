package usecase

import (
	"errors"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
)

func TestUserSiteUsecase_InsertUserSite(t *testing.T) {
	t.Run("should insert user site successfully", func(t *testing.T) {
		mockUserSiteRepo := new(mocks.MockUserSiteRepository)
		mockPlanRepo := new(mocks.MockPlanRepository)
		uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

		plan := &model.Plan{ID: 1, MaxSites: 5}
		mockPlanRepo.On("GetPlanByUserID", 1).Return(plan, nil).Once()
		mockUserSiteRepo.On("GetUserSiteCount", 1).Return(2, nil).Once()
		mockUserSiteRepo.On("InsertNewUserSite", 1, 10, []string{"golang"}).Return(nil).Once()

		err := uc.InsertUserSite(1, 10, []string{"golang"})

		assert.NoError(t, err)
		mockPlanRepo.AssertExpectations(t)
		mockUserSiteRepo.AssertExpectations(t)
	})

	t.Run("should return error when plan not found", func(t *testing.T) {
		mockUserSiteRepo := new(mocks.MockUserSiteRepository)
		mockPlanRepo := new(mocks.MockPlanRepository)
		uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

		mockPlanRepo.On("GetPlanByUserID", 1).Return(nil, nil).Once()

		err := uc.InsertUserSite(1, 10, []string{"golang"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nenhum plano associado")
		mockPlanRepo.AssertExpectations(t)
	})

	t.Run("should return error when site limit exceeded", func(t *testing.T) {
		mockUserSiteRepo := new(mocks.MockUserSiteRepository)
		mockPlanRepo := new(mocks.MockPlanRepository)
		uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

		plan := &model.Plan{ID: 1, MaxSites: 3}
		mockPlanRepo.On("GetPlanByUserID", 1).Return(plan, nil).Once()
		mockUserSiteRepo.On("GetUserSiteCount", 1).Return(3, nil).Once()

		err := uc.InsertUserSite(1, 10, []string{"golang"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limite de sites atingido")
		mockPlanRepo.AssertExpectations(t)
		mockUserSiteRepo.AssertExpectations(t)
	})

	t.Run("should return error when count fails", func(t *testing.T) {
		mockUserSiteRepo := new(mocks.MockUserSiteRepository)
		mockPlanRepo := new(mocks.MockPlanRepository)
		uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

		plan := &model.Plan{ID: 1, MaxSites: 5}
		mockPlanRepo.On("GetPlanByUserID", 1).Return(plan, nil).Once()
		mockUserSiteRepo.On("GetUserSiteCount", 1).Return(0, errors.New("db error")).Once()

		err := uc.InsertUserSite(1, 10, []string{"golang"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao contar sites")
		mockPlanRepo.AssertExpectations(t)
		mockUserSiteRepo.AssertExpectations(t)
	})

	t.Run("should return error when GetPlanByUserID fails", func(t *testing.T) {
		mockUserSiteRepo := new(mocks.MockUserSiteRepository)
		mockPlanRepo := new(mocks.MockPlanRepository)
		uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

		mockPlanRepo.On("GetPlanByUserID", 1).Return(nil, errors.New("plan db error")).Once()

		err := uc.InsertUserSite(1, 10, []string{"golang"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar plano")
		mockPlanRepo.AssertExpectations(t)
	})
}

func TestUserSiteUsecase_DeleteUserSite(t *testing.T) {
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockPlanRepo := new(mocks.MockPlanRepository)
	uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

	t.Run("should delete user site successfully", func(t *testing.T) {
		mockUserSiteRepo.On("DeleteUserSite", 1, "10").Return(nil).Once()

		err := uc.DeleteUserSite(1, "10")

		assert.NoError(t, err)
		mockUserSiteRepo.AssertExpectations(t)
	})
}

func TestUserSiteUsecase_UpdateUserSiteFilters(t *testing.T) {
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockPlanRepo := new(mocks.MockPlanRepository)
	uc := NewUserSiteUsecase(mockUserSiteRepo, mockPlanRepo)

	t.Run("should update filters successfully", func(t *testing.T) {
		mockUserSiteRepo.On("UpdateUserSiteFilters", 1, 10, []string{"go", "backend"}).Return(nil).Once()

		err := uc.UpdateUserSiteFilters(1, 10, []string{"go", "backend"})

		assert.NoError(t, err)
		mockUserSiteRepo.AssertExpectations(t)
	})
}
