package scrapper

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"web-scrapper/model"
	"github.com/gocolly/colly"
)


type JobScrapper struct{
}

func NewJobScraper() *JobScrapper {
	return &JobScrapper{
	}
}

func (s *JobScrapper) configureCollyCallbacks(c *colly.Collector, detailCollector *colly.Collector, jobs []*model.Job, wg *sync.WaitGroup, mu *sync.Mutex, selectors model.SiteScrapingConfig){

	nextPageVisitedOnThisRequest := true

	detailCollector.OnHTML("body", func(e *colly.HTMLElement) {
		defer wg.Done()

		jobPtr := e.Request.Ctx.GetAny("job").(*model.Job)

		descriptionHTML:= e.ChildText(*selectors.JobDescriptionSelector)
		if descriptionHTML == "" {
			log.Printf("Erro ao extrair HTML da descrição na vaga %s:", jobPtr.Title)
			jobPtr.Description = ""
		} else {
			jobPtr.Description = strings.TrimSpace(descriptionHTML)
		}
		jobIDstr := e.ChildText(*selectors.JobRequisitionIdSelector)
		jobID, err := strconv.Atoi(jobIDstr)
		log.Printf("job id for %s : %d", jobPtr.Title, jobID)
		if err != nil {
			fmt.Println("Erro ao converter ID:", err)
		} else {
			jobPtr.Requisition_ID = jobID
		}
	})

	c.OnHTML(*selectors.JobListItemSelector, func(e *colly.HTMLElement) {
		Title := e.ChildText(*selectors.TitleSelector)
		JobLink := e.ChildAttr(*selectors.LinkSelector, *selectors.LinkAttribute)
		Location := e.ChildText(*selectors.LocationSelector)

		job := &model.Job{
			Title:    Title,
			Location: Location,
			Job_link: JobLink,
		}

		if JobLink != "" {
			wg.Add(1)
			jobURL := e.Request.AbsoluteURL(JobLink)
			ctx := colly.NewContext()
			ctx.Put("job", job)
			detailCollector.Request("GET", jobURL, nil, ctx, nil)
		}
		mu.Lock()
		jobs = append(jobs, job)
		mu.Unlock()
	})

	c.OnHTML(*selectors.NextPageSelector, func(e *colly.HTMLElement) {
		if !nextPageVisitedOnThisRequest {
            nextPage := e.Request.AbsoluteURL(e.Attr("href"))
            if nextPage != "" {
                fmt.Printf("Visiting next page: %s\n", nextPage)
                nextPageVisitedOnThisRequest = true 
                e.Request.Visit(nextPage)
            }
        }
	})
}

func (s *JobScrapper) Scrape(ctx context.Context, config model.SiteScrapingConfig) ([]*model.Job, error) {
	var jobs []*model.Job
	var wg sync.WaitGroup
	var mu sync.Mutex

	c := colly.NewCollector(colly.Async(true))
	detailCollector := c.Clone()
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 8})
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"

	s.configureCollyCallbacks(c, detailCollector, jobs, &wg, &mu, config)

	if err := c.Visit(config.BaseURL); err != nil {
		return nil, err
	}

	c.Wait()
	wg.Wait()
	return jobs, nil
}