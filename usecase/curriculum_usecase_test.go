package usecase

import (
	"errors"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
)

func TestCurriculumUsecase_CreateCurriculum(t *testing.T) {
	mockRepo := new(mocks.MockCurriculumRepository)
	uc := NewCurriculumUsecase(mockRepo)

	t.Run("should create curriculum successfully", func(t *testing.T) {
		input := model.Curriculum{Title: "My CV", UserID: 1, Skills: "Go, React"}
		expected := model.Curriculum{Id: 1, Title: "My CV", UserID: 1, Skills: "Go, React"}

		mockRepo.On("CreateCurriculum", input).Return(expected, nil).Once()

		result, err := uc.CreateCurriculum(input)

		assert.NoError(t, err)
		assert.Equal(t, expected.Id, result.Id)
		assert.Equal(t, expected.Title, result.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repo fails", func(t *testing.T) {
		input := model.Curriculum{Title: "Fail CV", UserID: 2}

		mockRepo.On("CreateCurriculum", input).Return(model.Curriculum{}, errors.New("db error")).Once()

		_, err := uc.CreateCurriculum(input)

		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCurriculumUsecase_GetCurriculumByUserId(t *testing.T) {
	mockRepo := new(mocks.MockCurriculumRepository)
	uc := NewCurriculumUsecase(mockRepo)

	t.Run("should return curricula for user", func(t *testing.T) {
		expected := []model.Curriculum{
			{Id: 1, Title: "CV 1", UserID: 1},
			{Id: 2, Title: "CV 2", UserID: 1},
		}

		mockRepo.On("FindCurriculumByUserID", 1).Return(expected, nil).Once()

		result, err := uc.GetCurriculumByUserId(1)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return empty slice when no curricula", func(t *testing.T) {
		mockRepo.On("FindCurriculumByUserID", 99).Return([]model.Curriculum{}, nil).Once()

		result, err := uc.GetCurriculumByUserId(99)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurriculumUsecase_UpdateCurriculum(t *testing.T) {
	mockRepo := new(mocks.MockCurriculumRepository)
	uc := NewCurriculumUsecase(mockRepo)

	t.Run("should update curriculum successfully", func(t *testing.T) {
		input := model.Curriculum{Id: 1, Title: "Updated CV", UserID: 1, Skills: "Go, Python"}
		expected := model.Curriculum{Id: 1, Title: "Updated CV", UserID: 1, Skills: "Go, Python"}

		mockRepo.On("UpdateCurriculum", input).Return(expected, nil).Once()

		result, err := uc.UpdateCurriculum(input)

		assert.NoError(t, err)
		assert.Equal(t, "Updated CV", result.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repo fails", func(t *testing.T) {
		input := model.Curriculum{Id: 999, Title: "Bad CV"}

		mockRepo.On("UpdateCurriculum", input).Return(model.Curriculum{}, errors.New("not found")).Once()

		_, err := uc.UpdateCurriculum(input)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCurriculumUsecase_DeleteCurriculum(t *testing.T) {
	mockRepo := new(mocks.MockCurriculumRepository)
	uc := NewCurriculumUsecase(mockRepo)

	t.Run("should delete curriculum successfully when user has multiple", func(t *testing.T) {
		mockRepo.On("DeleteCurriculumIfNotLast", 1, 5).Return(nil).Once()

		err := uc.DeleteCurriculum(1, 5)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when only one curriculum", func(t *testing.T) {
		mockRepo.On("DeleteCurriculumIfNotLast", 1, 5).Return(errors.New("não é possível excluir o único currículo")).Once()

		err := uc.DeleteCurriculum(1, 5)

		assert.Error(t, err)
		assert.Equal(t, "não é possível excluir o único currículo", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repo fails", func(t *testing.T) {
		mockRepo.On("DeleteCurriculumIfNotLast", 1, 5).Return(errors.New("db error")).Once()

		err := uc.DeleteCurriculum(1, 5)

		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
