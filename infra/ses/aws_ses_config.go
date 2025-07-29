package ses

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func LoadAWSConfig(ct context.Context) (aws.Config, error){
	if endpointURL := os.Getenv("AWS_ENDPOINT_URL"); endpointURL != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           endpointURL,
				SigningRegion: "us-east-1",
			}, nil
		})

		return config.LoadDefaultConfig(ct, config.WithEndpointResolverWithOptions(customResolver))
	}

	return config.LoadDefaultConfig(ct)
}

func LoadAWSClient(cfg aws.Config) (*ses.Client){
	return ses.NewFromConfig(cfg)
}