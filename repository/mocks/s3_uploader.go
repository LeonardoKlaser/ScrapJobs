package mocks

import (
	"context"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

type MockS3Uploader struct {
	mock.Mock
}

func (m *MockS3Uploader) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, file)
	return args.String(0), args.Error(1)
}
