package scrapper

import (
	"sync"
	"bytes"
	"context"
	"fmt"
	"log"
	"web-scrapper/model"
	"github.com/chromedp/chromedp"
    "github.com/gocolly/colly"
)


type HeadlessScraper struct{
	collyParser *JobScrapper
}

func NewHeadlessScraper() *HeadlessScraper {
	return &HeadlessScraper{
		collyParser : NewJobScraper(),
	}
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

	var htmlContent string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(config.BaseURL),
		chromedp.WaitVisible(config.JobListItemSelector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, fmt.Errorf("ERROR chrome automation fail to %s: %w", config.SiteName, err)
	}

	if htmlContent == "" {
        return nil, fmt.Errorf("error to remain HTML content from page %s", config.SiteName)
    }

	var jobs []*model.Job
	var wg sync.WaitGroup
	var mu sync.Mutex

	c := colly.NewCollector(colly.Async(true))
	detailCollector := c.Clone()
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 8})

	s.collyParser.configureCollyCallbacks(c, detailCollector, &jobs, &wg, &mu, config)

	err = c.Request("GET", config.BaseURL, bytes.NewBufferString(htmlContent), nil, nil)
    if err != nil {
        return nil, fmt.Errorf("colly falhou ao processar o HTML do chromedp: %w", err)
    }

	c.Wait()
	wg.Wait()

	return jobs, nil 
}