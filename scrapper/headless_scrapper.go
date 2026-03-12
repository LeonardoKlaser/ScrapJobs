package scrapper

import (
	"context"
	"fmt"
	"net/url"
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
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
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
	sem := make(chan struct{}, 3) // max 3 concurrent detail page fetches

	parsedBaseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL %s: %w", config.BaseURL, err)
	}

	doc.Find(*config.JobListItemSelector).Each(func(_ int, sel *goquery.Selection) {
		title := strings.TrimSpace(sel.Find(*config.TitleSelector).Text())

		var jobLink string
		if config.LinkSelector != nil && config.LinkAttribute != nil {
			rawLink, _ := sel.Find(*config.LinkSelector).Attr(*config.LinkAttribute)
			if rawLink != "" {
				parsedLink, err := url.Parse(rawLink)
				if err == nil {
					jobLink = parsedBaseURL.ResolveReference(parsedLink).String()
				} else {
					jobLink = rawLink
				}
			}
		}

		if jobLink != "" && !strings.HasPrefix(jobLink, "http://") && !strings.HasPrefix(jobLink, "https://") {
			if config.APIEndpointTemplate != nil && *config.APIEndpointTemplate != "" {
				jobLink = *config.APIEndpointTemplate + jobLink
			}
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

		if jobLink != "" && (config.JobDescriptionSelector != nil || config.JobRequisitionIdSelector != nil) {
			wg.Add(1)
			sem <- struct{}{}
			go func(j *model.Job, link string) {
				defer wg.Done()
				defer func() { <-sem }()
				s.fetchJobDetails(allocCtx, config, j, link)
			}(job, jobLink)
		}

		mu.Lock()
		jobs = append(jobs, job)
		mu.Unlock()
	})

	wg.Wait()
	return jobs, nil
}

func (s *HeadlessScraper) fetchJobDetails(allocCtx context.Context, config model.SiteScrapingConfig, job *model.Job, jobURL string) {
	taskCtx, cancel := chromedp.NewContext(allocCtx) // reuses existing Chrome allocator
	defer cancel()

	// Use smarter wait: prefer the description selector if available, otherwise just wait for body
	var waitAction chromedp.Action
	if config.JobDescriptionSelector != nil {
		waitAction = chromedp.WaitVisible(*config.JobDescriptionSelector, chromedp.ByQuery)
	} else {
		waitAction = chromedp.WaitReady("body")
	}

	var detailHTML string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(jobURL),
		waitAction,
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

	if config.JobRequisitionIdSelector != nil {
		reqID := strings.TrimSpace(detailDoc.Find(*config.JobRequisitionIdSelector).Text())
		if reqID == "" {
			logging.Logger.Warn().Str("job_title", job.Title).Msg("Failed to extract requisition ID")
		} else {
			logging.Logger.Debug().Str("job_title", job.Title).Str("requisition_id", reqID).Msg("Parsed requisition ID")
			job.RequisitionID = reqID
		}
	}
}
