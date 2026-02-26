package usecase

import (
	"errors"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
)

func TestJobUseCase_CreateJob(t *testing.T) {
	mockRepo := new(mocks.MockJobRepository)
	uc := NewJobUseCase(mockRepo)

	t.Run("should create job successfully", func(t *testing.T) {
		job := model.Job{
			SiteID:        1,
			Title:         "Go Developer",
			Location:      "Remote",
			Company:       "Acme",
			JobLink:       "https://acme.com/jobs/1",
			RequisitionID: 12345,
		}

		mockRepo.On("CreateJob", job).Return(1, nil).Once()

		id, err := uc.CreateJob(job)

		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repo fails", func(t *testing.T) {
		job := model.Job{Title: "Fail Job"}

		mockRepo.On("CreateJob", job).Return(0, errors.New("duplicate key")).Once()

		_, err := uc.CreateJob(job)

		assert.Error(t, err)
		assert.Equal(t, "duplicate key", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestJobUseCase_FindJobByRequisitionID(t *testing.T) {
	mockRepo := new(mocks.MockJobRepository)
	uc := NewJobUseCase(mockRepo)

	t.Run("should return true when job exists", func(t *testing.T) {
		mockRepo.On("FindJobByRequisitionID", 12345).Return(true, nil).Once()

		found, err := uc.FindJobByRequisitionID(12345)

		assert.NoError(t, err)
		assert.True(t, found)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when job not found", func(t *testing.T) {
		mockRepo.On("FindJobByRequisitionID", 99999).Return(false, nil).Once()

		found, err := uc.FindJobByRequisitionID(99999)

		assert.NoError(t, err)
		assert.False(t, found)
		mockRepo.AssertExpectations(t)
	})
}
