package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func LoadAWSConfig(ct context.Context) (aws.Config, error){
	return config.LoadDefaultConfig(ct)
}

func LoadAWSClient(cfg aws.Config) (*ses.Client){
	return ses.NewFromConfig(cfg)
}