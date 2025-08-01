package interfaces

import(
	"context"
	"web-scrapper/model"
)

 type Scraper interface {
	Scrape(ctx context.Context, config model.SiteScrapingConfig) ([]*model.Job, error)
}