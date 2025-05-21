package ses

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func LoadAWSConfig(ct context.Context) (aws.Config, error){
	return config.LoadDefaultConfig(ct)
}