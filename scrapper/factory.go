
package scrapper

import (
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/model"
)

func NewScraperFactory(config model.SiteScrapingConfig) (interfaces.Scraper, error) {
	switch config.ScrapingType {
	case "HTML":
		return NewJobScraper(), nil 
	case "API":
		return NewAPIScrapper(), nil
	case "HEADLESS":
		return NewHeadlessScraper(), nil
	default:
		return nil, fmt.Errorf("scrap strategy not found: %s", config.ScrapingType)
	}
}