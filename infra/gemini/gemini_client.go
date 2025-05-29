package gemini 
import (
	"google.golang.org/genai"
	"context"
	"errors"
)

type Config struct {
	ApiKey string
	ApiModel string
}

type GeminiClient struct{
	genAIClient *genai.Client
	model string
}

func GeminiClientModel(ctx context.Context, cfg Config) (*GeminiClient, error) {
	if cfg.ApiKey == "" {
		return nil, errors.New("gemini API key is required")
	}
	if cfg.ApiModel == "" {
		return nil, errors.New("gemini model name is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.ApiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &GeminiClient{
		genAIClient: client,
		model: cfg.ApiModel,
	}, nil
}

func (c *GeminiClient)GeminiSearch( ctx context.Context, search string) (*genai.GenerateContentResponse, error){
	if c.genAIClient == nil {
		return nil, errors.New("gemini client not initialized")
	}

	parts := []*genai.Part{
		{Text: search},
	}

	result , err := c.genAIClient.Models.GenerateContent(ctx, c.model, []*genai.Content{{Parts: parts}}, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

