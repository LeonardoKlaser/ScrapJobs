package utils

import (
	"context"
	"encoding/json"
	"web-scrapper/model"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetAppSecrets(secretName string) (*model.AppSecrets, error){
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil{
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := svc.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, err
	}

	var secrets model.AppSecrets
	err = json.Unmarshal([]byte(*result.SecretString), &secrets)
	if err != nil {
		return nil, err
	}

	return &secrets, nil
}

