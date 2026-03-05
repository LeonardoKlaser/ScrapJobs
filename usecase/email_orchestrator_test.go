package usecase

import (
	"context"
	"errors"
	"testing"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailOrchestrator_SendEmail_PrimarySuccess(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: true, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockResend.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(nil)

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.NoError(t, err)
	mockResend.AssertExpectations(t)
	mockSES.AssertNotCalled(t, "SendEmail")
}

func TestEmailOrchestrator_SendEmail_FallbackOnPrimaryFailure(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: true, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockResend.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(errors.New("resend down"))
	mockSES.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(nil)

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.NoError(t, err)
	mockResend.AssertExpectations(t)
	mockSES.AssertExpectations(t)
}

func TestEmailOrchestrator_SendEmail_SkipsInactiveProvider(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: false, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockSES.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(nil)

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.NoError(t, err)
	mockResend.AssertNotCalled(t, "SendEmail")
	mockSES.AssertExpectations(t)
}

func TestEmailOrchestrator_SendEmail_AllProvidersFail(t *testing.T) {
	mockRepo := new(mocks.MockEmailConfigRepo)
	mockResend := new(mocks.MockMailSender)
	mockSES := new(mocks.MockMailSender)

	mockRepo.On("GetAll").Return([]model.EmailProviderConfig{
		{ProviderName: "resend", IsActive: true, Priority: 1},
		{ProviderName: "ses", IsActive: true, Priority: 2},
	}, nil)

	mockResend.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(errors.New("resend down"))
	mockSES.On("SendEmail", mock.Anything, "to@test.com", "subject", "text", "html").Return(errors.New("ses down"))

	senders := map[string]interfaces.MailSender{
		"resend": mockResend,
		"ses":    mockSES,
	}

	orch := NewEmailOrchestrator(senders, mockRepo)
	err := orch.SendEmail(context.Background(), "to@test.com", "subject", "text", "html")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all email providers failed")
}
