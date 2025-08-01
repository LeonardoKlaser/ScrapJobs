package scrapper

import (
	"context"
	"fmt"
	"log"
	"web-scrapper/model"

	"github.com/chromedp/chromedp"
)


type HeadlessScraper struct{}

func NewHeadlessScraper() *HeadlessScraper {
	return &HeadlessScraper{}
}


func (s *HeadlessScraper) Scrape(ctx context.Context, config model.SiteScrapingConfig) ([]*model.Job, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var jobsHTML string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(config.BaseURL),
		chromedp.WaitVisible(config.JobListItemSelector.String, chromedp.ByQuery),
		chromedp.OuterHTML(config.JobListItemSelector.String, &jobsHTML, chromedp.ByQuery),
	)

	if err != nil {
		return nil, fmt.Errorf("ERROR chrome automation fail to %s: %w", config.SiteName, err)
	}


	return []*model.Job{}, nil 
}