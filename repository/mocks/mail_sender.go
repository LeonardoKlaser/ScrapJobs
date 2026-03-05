package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockMailSender struct {
	mock.Mock
}

func (m *MockMailSender) SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error {
	args := m.Called(ctx, to, subject, bodyText, bodyHtml)
	return args.Error(0)
}
