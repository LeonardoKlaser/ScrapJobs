package scrapper

import (
	"fmt"
	"strconv"
	"strings"
	"web-scrapper/model"

	"github.com/gocolly/colly"
)

type JobScrapper interface {
	ScrapeJobs() ([]*model.Job, error)
}

type jobScraper struct{}

func NewJobScraper() JobScrapper {
	return &jobScraper{}
}

func (s *jobScraper) ScrapeJobs() ([]*model.Job, error) {
	var jobs []*model.Job
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	detailCollector := c.Clone()

	detailCollector.OnHTML("body", func(e *colly.HTMLElement) {
		jobPtr := e.Request.Ctx.GetAny("job").(*model.Job)
		jobIDstr := e.ChildText("span[data-careersite-propertyid=facility]")
		jobID, err := strconv.Atoi(jobIDstr)
		if err != nil {
			fmt.Println("Erro ao converter ID:", err)
		} else {
			jobPtr.Requisition_ID = jobID
		}
	})

	c.OnHTML("table#searchresults tr.data-row", func(e *colly.HTMLElement) {
		Title := e.ChildText("span.jobTitle.hidden-phone a.jobTitle-link")
		JobLink := e.ChildAttr("span.jobTitle.hidden-phone a.jobTitle-link", "href")
		Location := e.ChildText("td.colLocation span.jobLocation")
		target := "developer"
		target2 := "software"

		if strings.Contains(strings.ToUpper(Title), strings.ToUpper(target)) ||
			strings.Contains(strings.ToUpper(Title), strings.ToUpper(target2)) {
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
		}
	})

	c.OnHTML("a.paginationItemLast", func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPage != "" {
			fmt.Printf("Visiting next page: %s\n", nextPage)
			e.Request.Visit(nextPage)
		}
	})

	err := c.Visit("https://jobs.sap.com/search/?q=&locationsearch=S%C3%A3o+Leopoldo&location=S%C3%A3o+Leopoldo&scrollToTable=true")
	if err != nil {
		return nil, err
	}

	return jobs, nil
}