package scrapper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"web-scrapper/logging"
	"web-scrapper/model"

	"github.com/PuerkitoBio/goquery"
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

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
		logging.Logger.Debug().Msgf(format, args...)
	}))
	defer cancel()

	if config.JobListItemSelector == nil || config.TitleSelector == nil || config.LinkSelector == nil || config.LinkAttribute == nil {
		return nil, fmt.Errorf("required selectors (JobListItemSelector, TitleSelector, LinkSelector, LinkAttribute) must not be nil for headless scraping of %s", config.SiteName)
	}

	var htmlContent string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(config.BaseURL),
		chromedp.WaitVisible(*config.JobListItemSelector, chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, fmt.Errorf("ERROR chrome automation fail to %s: %w", config.SiteName, err)
	}

	if htmlContent == "" {
		return nil, fmt.Errorf("error to remain HTML content from page %s", config.SiteName)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error parsing rendered HTML for %s: %w", config.SiteName, err)
	}

	var jobs []*model.Job
	var mu sync.Mutex
	var wg sync.WaitGroup

	doc.Find(*config.JobListItemSelector).Each(func(_ int, sel *goquery.Selection) {
		title := strings.TrimSpace(sel.Find(*config.TitleSelector).Text())

		var jobLink string
		if config.LinkSelector != nil && config.LinkAttribute != nil {
			jobLink, _ = sel.Find(*config.LinkSelector).Attr(*config.LinkAttribute)
		}

		var location string
		if config.LocationSelector != nil {
			location = strings.TrimSpace(sel.Find(*config.LocationSelector).Text())
		}

		job := &model.Job{
			Title:    title,
			Location: location,
			JobLink:  jobLink,
		}

		if jobLink != "" && config.JobDescriptionSelector != nil {
			wg.Add(1)
			go func(j *model.Job, link string) {
				defer wg.Done()
				s.fetchJobDetails(ctx, config, j, link)
			}(job, jobLink)
		}

		mu.Lock()
		jobs = append(jobs, job)
		mu.Unlock()
	})

	wg.Wait()
	return jobs, nil
}

func (s *HeadlessScraper) fetchJobDetails(ctx context.Context, config model.SiteScrapingConfig, job *model.Job, jobURL string) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var detailHTML string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(jobURL),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("body", &detailHTML),
	)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("job_title", job.Title).Str("url", jobURL).Msg("Failed to fetch job detail page")
		return
	}

	detailDoc, err := goquery.NewDocumentFromReader(strings.NewReader(detailHTML))
	if err != nil {
		logging.Logger.Warn().Err(err).Str("job_title", job.Title).Msg("Failed to parse job detail HTML")
		return
	}

	if config.JobDescriptionSelector != nil {
		description := strings.TrimSpace(detailDoc.Find(*config.JobDescriptionSelector).Text())
		job.Description = description
	}
}
