package usecase

import (
	"fmt"
	"testing"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
)

func TestRequestedSiteUsecase_Create_Success(t *testing.T) {
	mockRepo := new(mocks.MockRequestedSiteRepository)
	uc := NewRequestedSiteUsecase(mockRepo)

	mockRepo.On("Create", 1, "https://careers.example.com").Return(nil)

	err := uc.Create(1, "https://careers.example.com")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRequestedSiteUsecase_Create_RepoError(t *testing.T) {
	mockRepo := new(mocks.MockRequestedSiteRepository)
	uc := NewRequestedSiteUsecase(mockRepo)

	mockRepo.On("Create", 1, "https://careers.example.com").Return(fmt.Errorf("database error"))

	err := uc.Create(1, "https://careers.example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}
