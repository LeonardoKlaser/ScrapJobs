package ses

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func LoadAWSConfig(ct context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(ct)
}

func LoadAWSClient(cfg aws.Config) *ses.Client {
	if endpointURL := os.Getenv("AWS_ENDPOINT_URL"); endpointURL != "" {
		return ses.NewFromConfig(cfg, func(o *ses.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		})
	}
	return ses.NewFromConfig(cfg)
}
