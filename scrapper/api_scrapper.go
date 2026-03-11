package scrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	"web-scrapper/model"

	"github.com/tidwall/gjson"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

	return s.parseAPIResponse(body, *config.JSONDataMappings, config.BaseURL)
}

type Mapeamentos struct {
	JobsArrayPath     string `json:"jobs_array_path"`
	TitlePath         string `json:"title_path"`
	LinkPath          string `json:"link_path"`
	LocationPath      string `json:"location_path"`
	DescriptionPath   string `json:"description_path"`
	RequisitionIDPath string `json:"requisition_id_path"`
}

func (s *APIScrapper) parseAPIResponse(body []byte, mappingsJSON string, baseURL string) ([]*model.Job, error) {
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
		title := value.Get(mappings.TitlePath).String()
		jobLink := value.Get(mappings.LinkPath).String()
		if jobLink != "" && !strings.HasPrefix(jobLink, "http") {
			slug := generateSlug(title)
			base := strings.TrimRight(baseURL, "/")
			id := strings.TrimLeft(jobLink, "/")
			if slug != "" {
				jobLink = base + "/" + id + "/" + slug
			} else {
				jobLink = base + "/" + id
			}
		}

		job := &model.Job{
			Title:       title,
			JobLink:     jobLink,
			Location:    value.Get(mappings.LocationPath).String(),
			Description: value.Get(mappings.DescriptionPath).String(),
		}

		reqIDStr := value.Get(mappings.RequisitionIDPath).String()
		if reqIDStr != "" {
			job.RequisitionID = reqIDStr
		}

		jobs = append(jobs, job)
		return true 
	})

	return jobs, nil
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(title string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, title)
	if err != nil {
		result = title
	}
	result = strings.ToLower(result)
	result = nonAlphanumeric.ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")
	return result
}