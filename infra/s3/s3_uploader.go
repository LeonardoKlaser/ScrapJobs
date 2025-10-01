package s3

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Uploader struct {
	Client     *s3.Client
	BucketName string
}

func NewUploader(cfg aws.Config, bucketName string) *Uploader {
	client := s3.NewFromConfig(cfg)
	return &Uploader{
		Client:     client,
		BucketName: bucketName,
	}
}

func (u *Uploader) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("falha ao abrir o arquivo: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	key := fmt.Sprintf("logos/%s%s", uuid.New().String(), ext)

	_, err = u.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.BucketName),
		Key:         aws.String(key),
		Body:        src,
		ACL:         "public-read",
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", fmt.Errorf("falha ao fazer upload para o S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.BucketName, key)

	return url, nil
}

type UploaderInterface interface {
    UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error)
}