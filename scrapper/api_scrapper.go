package scrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"web-scrapper/model"

	"github.com/tidwall/gjson"
)

type APIScrapper struct {
	client *http.Client
}

func NewAPIScrapper() *APIScrapper{
	return &APIScrapper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *APIScrapper) Scrape(ctx context.Context, config model.SiteScrapingConfig) ([]*model.Job, error){
	if config.APIEndpointTemplate == nil {
		return nil, fmt.Errorf("API endpoint template is required for site %s", config.SiteName)
	}
	if config.JSONDataMappings == nil {
		return nil, fmt.Errorf("JSON data mappings is required for site %s", config.SiteName)
	}

	method := "GET"
	if config.APIMethod != nil && *config.APIMethod != "" {
		method = *config.APIMethod
	}

	req, err := http.NewRequestWithContext(ctx, method, *config.APIEndpointTemplate, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR to create request %s: %w", config.SiteName, err)
	}

	if config.APIHeadersJSON != nil && *config.APIHeadersJSON != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(*config.APIHeadersJSON), &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR to execute request %s: %w", config.SiteName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return nil, fmt.Errorf("unexpected status code to %s: %d", config.SiteName, resp.StatusCode)
	}

	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler corpo da resposta de %s: %w", config.SiteName, err)
	}

	return s.parseAPIResponse(body, *config.JSONDataMappings)
}

type Mapeamentos struct {
	JobsArrayPath     string `json:"jobs_array_path"`
	TitlePath         string `json:"title_path"`
	LinkPath          string `json:"link_path"`
	LocationPath      string `json:"location_path"`
	DescriptionPath   string `json:"description_path"`
	RequisitionIDPath string `json:"requisition_id_path"`
}

func (s *APIScrapper) parseAPIResponse(body []byte, mappingsJSON string) ([]*model.Job, error) {
	var mappings Mapeamentos
	if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
		return nil, fmt.Errorf("ERROR to parse json maps: %w", err)
	}

	var jobs []*model.Job
	result := gjson.Get(string(body), mappings.JobsArrayPath)

	if !result.Exists() {
		return nil, fmt.Errorf("array path not found: %s", mappings.JobsArrayPath)
	}

	result.ForEach(func(key, value gjson.Result) bool {
		job := &model.Job{
			Title:       value.Get(mappings.TitlePath).String(),
			JobLink:    value.Get(mappings.LinkPath).String(),
			Location:    value.Get(mappings.LocationPath).String(),
			Description: value.Get(mappings.DescriptionPath).String(),
		}

		reqIDStr := value.Get(mappings.RequisitionIDPath).String()
		if reqID, err := strconv.Atoi(reqIDStr); err == nil {
			job.RequisitionID = int64(reqID)
		}

		jobs = append(jobs, job)
		return true 
	})

	return jobs, nil
}