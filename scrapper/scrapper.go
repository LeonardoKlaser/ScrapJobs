package scrapper

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"web-scrapper/model"
	"github.com/gocolly/colly"
)

type JobScrapper interface {
	ScrapeJobs(selectors model.SiteScrapingConfig) ([]*model.Job, error)
}

type jobScraper struct{
}

func NewJobScraper() JobScrapper {
	return &jobScraper{
	}
}

func (s *jobScraper) ScrapeJobs(selectors model.SiteScrapingConfig) ([]*model.Job, error) {
	var jobs []*model.Job
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	detailCollector := c.Clone()

	detailCollector.OnHTML("body", func(e *colly.HTMLElement) {
		jobPtr := e.Request.Ctx.GetAny("job").(*model.Job)

		descriptionHTML:= e.ChildText(selectors.JobDescriptionSelector)
		if descriptionHTML == "" {
			fmt.Println("Erro ao extrair HTML da descrição:")
			jobPtr.Description = ""
		} else {
			jobPtr.Description = strings.TrimSpace(descriptionHTML)
		}
		jobIDstr := e.ChildText(selectors.JobRequisitionIdSelector)
		jobID, err := strconv.Atoi(jobIDstr)
		if err != nil {
			fmt.Println("Erro ao converter ID:", err)
		} else {
			jobPtr.Requisition_ID = jobID
		}
	})

	c.OnHTML(selectors.JobListItemSelector, func(e *colly.HTMLElement) {
		Title := e.ChildText(selectors.TitleSelector)
		JobLink := e.ChildAttr(selectors.LinkSelector, "href")
		Location := e.ChildText(selectors.LocationSelector)

		job := &model.Job{
			Title:    Title,
			Location: Location,
			Job_link: JobLink,
		}

		if JobLink != "" {
			jobURL := e.Request.AbsoluteURL(JobLink)
			ctx := colly.NewContext()
			ctx.Put("job", job)
			detailCollector.Request("GET", jobURL, nil, ctx, nil)
		}

		jobs = append(jobs, job)
	})

	c.OnHTML(selectors.NextPageSelector, func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPage != "" {
			fmt.Printf("Visiting next page: %s\n", nextPage)
			e.Request.Visit(nextPage)
		}
	})

	err := c.Visit(selectors.BaseURL)
	if err != nil {
		return nil, err
	}
	log.Printf("Retornando: %v", jobs)
	return jobs, nil
}