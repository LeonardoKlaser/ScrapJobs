package scrapper

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"web-scrapper/logging"
	"web-scrapper/model"

	"github.com/gocolly/colly/v2"
)

type JobScrapper struct{
}

func NewJobScraper() *JobScrapper {
	return &JobScrapper{
	}
}

func (s *JobScrapper) configureCollyCallbacks(c *colly.Collector, detailCollector *colly.Collector, jobs *[]*model.Job, wg *sync.WaitGroup, mu *sync.Mutex, selectors model.SiteScrapingConfig){

	detailCollector.OnHTML("body", func(e *colly.HTMLElement) {
		defer wg.Done()

		jobPtr := e.Request.Ctx.GetAny("job").(*model.Job)

		if selectors.JobDescriptionSelector != nil{
			descriptionHTML:= e.ChildText(*selectors.JobDescriptionSelector)
			jobPtr.Description = strings.TrimSpace(descriptionHTML)

			if descriptionHTML == "" {
				logging.Logger.Warn().Str("job_title", jobPtr.Title).Msg("Failed to extract description HTML")
				jobPtr.Description = ""
			}
		}

		if selectors.JobRequisitionIdSelector != nil {
			jobIDstr := e.ChildText(*selectors.JobRequisitionIdSelector)
			jobID, err := strconv.Atoi(jobIDstr)
			if err != nil {
				logging.Logger.Warn().Err(err).Str("job_title", jobPtr.Title).Msg("Failed to convert requisition ID")
			} else {
				logging.Logger.Debug().Str("job_title", jobPtr.Title).Int("requisition_id", jobID).Msg("Parsed requisition ID")
				jobPtr.RequisitionID = int64(jobID)
			}
		}
	})

	c.OnHTML(*selectors.JobListItemSelector, func(e *colly.HTMLElement) {
		Title := e.ChildText(*selectors.TitleSelector)
		JobLink := e.ChildAttr(*selectors.LinkSelector, *selectors.LinkAttribute)

		var Location string
		if selectors.LocationSelector != nil {
			Location = e.ChildText(*selectors.LocationSelector)
		}

		job := &model.Job{
			Title:    Title,
			Location: Location,
			JobLink: JobLink,
		}

		if JobLink != "" {
			wg.Add(1)
			jobURL := e.Request.AbsoluteURL(JobLink)
			ctx := colly.NewContext()
			ctx.Put("job", job)
			detailCollector.Request("GET", jobURL, nil, ctx, nil)
		}
		mu.Lock()
		*jobs = append(*jobs, job)
		mu.Unlock()
	})

	if selectors.NextPageSelector != nil {
		c.OnHTML(*selectors.NextPageSelector, func(e *colly.HTMLElement) {
			if e.Request.Ctx.Get("visitNextPage") == "true" {
				nextPage := e.Request.AbsoluteURL(e.Attr("href"))
				if nextPage != "" {
					logging.Logger.Debug().Str("url", nextPage).Msg("Visiting next page")
					e.Request.Ctx.Put("visitNextPage", "false")
					e.Request.Visit(nextPage)
				}
			}
		})
	}
}

func (s *JobScrapper) Scrape(ctx context.Context, config model.SiteScrapingConfig) ([]*model.Job, error) {
	if config.JobListItemSelector == nil || config.TitleSelector == nil || config.LinkSelector == nil || config.LinkAttribute == nil {
		return nil, fmt.Errorf("required selectors (JobListItemSelector, TitleSelector, LinkSelector, LinkAttribute) must not be nil for site %s", config.SiteName)
	}

	var jobs []*model.Job
	var wg sync.WaitGroup
	var mu sync.Mutex

	c := colly.NewCollector(colly.Async(true))
	detailCollector := c.Clone()
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 8})
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"

	c.OnRequest(func(r *colly.Request){
		r.Ctx.Put("visitNextPage", "true")
	})

	s.configureCollyCallbacks(c, detailCollector, &jobs, &wg, &mu, config)

	done := make(chan error, 1)
	go func() {
		if err := c.Visit(config.BaseURL); err != nil {
			done <- err
			return
		}
		c.Wait()
		wg.Wait()
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("scraping timed out for site %s: %w", config.SiteName, ctx.Err())
	case err := <-done:
		if err != nil {
			return nil, err
		}
		return jobs, nil
	}
}
