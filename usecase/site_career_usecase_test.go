package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSiteCareerUsecase_GetAllSites(t *testing.T) {
	mockRepo := new(mocks.MockSiteCareerRepository)
	mockUploader := new(mocks.MockS3Uploader)
	uc := NewSiteCareerUsecase(mockRepo, mockUploader)

	t.Run("should return all sites", func(t *testing.T) {
		expected := []model.SiteScrapingConfig{
			{ID: 1, SiteName: "Acme", BaseURL: "https://acme.com/careers", IsActive: true},
			{ID: 2, SiteName: "Beta", BaseURL: "https://beta.com/jobs", IsActive: false},
		}

		mockRepo.On("GetAllSites").Return(expected, nil).Once()

		result, err := uc.GetAllSites()

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Acme", result[0].SiteName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repo fails", func(t *testing.T) {
		mockRepo.On("GetAllSites").Return([]model.SiteScrapingConfig(nil), errors.New("db error")).Once()

		_, err := uc.GetAllSites()

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSiteCareerUsecase_InsertNewSiteCareer(t *testing.T) {
	t.Run("should insert site without logo", func(t *testing.T) {
		mockRepo := new(mocks.MockSiteCareerRepository)
		mockUploader := new(mocks.MockS3Uploader)
		uc := NewSiteCareerUsecase(mockRepo, mockUploader)

		site := model.SiteScrapingConfig{SiteName: "Acme", BaseURL: "https://acme.com", ScrapingType: "CSS"}
		expected := model.SiteScrapingConfig{ID: 1, SiteName: "Acme", BaseURL: "https://acme.com", ScrapingType: "CSS"}

		mockRepo.On("InsertNewSiteCareer", site).Return(expected, nil).Once()

		result, err := uc.InsertNewSiteCareer(context.Background(), site, nil)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.ID)
		mockUploader.AssertNotCalled(t, "UploadFile")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should insert site with logo upload", func(t *testing.T) {
		mockRepo := new(mocks.MockSiteCareerRepository)
		mockUploader := new(mocks.MockS3Uploader)
		uc := NewSiteCareerUsecase(mockRepo, mockUploader)

		site := model.SiteScrapingConfig{SiteName: "Beta", BaseURL: "https://beta.com", ScrapingType: "API"}
		file := &multipart.FileHeader{Filename: "logo.png", Size: 1024}

		logoURL := "https://bucket.s3.amazonaws.com/logos/uuid.png"
		mockUploader.On("UploadFile", mock.Anything, file).Return(logoURL, nil).Once()

		siteWithLogo := site
		siteWithLogo.LogoURL = &logoURL
		expected := model.SiteScrapingConfig{ID: 2, SiteName: "Beta", BaseURL: "https://beta.com", ScrapingType: "API", LogoURL: &logoURL}

		mockRepo.On("InsertNewSiteCareer", mock.MatchedBy(func(s model.SiteScrapingConfig) bool {
			return s.SiteName == "Beta" && s.LogoURL != nil && *s.LogoURL == logoURL
		})).Return(expected, nil).Once()

		result, err := uc.InsertNewSiteCareer(context.Background(), site, file)

		assert.NoError(t, err)
		assert.Equal(t, 2, result.ID)
		assert.NotNil(t, result.LogoURL)
		mockUploader.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when S3 upload fails", func(t *testing.T) {
		mockRepo := new(mocks.MockSiteCareerRepository)
		mockUploader := new(mocks.MockS3Uploader)
		uc := NewSiteCareerUsecase(mockRepo, mockUploader)

		site := model.SiteScrapingConfig{SiteName: "Fail", BaseURL: "https://fail.com"}
		file := &multipart.FileHeader{Filename: "big.png", Size: 5000000}

		mockUploader.On("UploadFile", mock.Anything, file).Return("", errors.New("file too large")).Once()

		_, err := uc.InsertNewSiteCareer(context.Background(), site, file)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "falha no upload do logo")
		mockRepo.AssertNotCalled(t, "InsertNewSiteCareer")
		mockUploader.AssertExpectations(t)
	})

	t.Run("should return error when repo fails after upload", func(t *testing.T) {
		mockRepo := new(mocks.MockSiteCareerRepository)
		mockUploader := new(mocks.MockS3Uploader)
		uc := NewSiteCareerUsecase(mockRepo, mockUploader)

		site := model.SiteScrapingConfig{SiteName: "RepoFail", BaseURL: "https://repofail.com"}

		mockRepo.On("InsertNewSiteCareer", site).Return(model.SiteScrapingConfig{}, errors.New("db error")).Once()

		_, err := uc.InsertNewSiteCareer(context.Background(), site, nil)

		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
