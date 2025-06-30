package usecase

import (
	"context"
	"log"
	"web-scrapper/model"
	"web-scrapper/repository"
)

type SiteScraper interface {
    ScrapeAndStoreJobs(ctx context.Context, site model.SiteScrapingConfig) ([]*model.Job, error)
}

type MatchFinderNotifier interface {
    FindMatchesAndNotify(siteID int, jobs []*model.Job) error
}

type ScrapingOrchestrator struct{
	siteRepo repository.SiteCareerRepository
	scraper SiteScraper
	notifier MatchFinderNotifier
}

func NewScrapingOrchestrator(
	siteRepo repository.SiteCareerRepository,
	scraper SiteScraper,
	notifier MatchFinderNotifier) *ScrapingOrchestrator{
		return &ScrapingOrchestrator{
			siteRepo:  siteRepo,
			scraper:   scraper,
			notifier:  notifier,
		}
}

func (o *ScrapingOrchestrator) ExecuteScrapingCycle(ctx context.Context){
	sites, err := o.siteRepo.GetAllSites()
	if err != nil {
		log.Printf("error to get all sites: %v", err)
	}

	for _, site := range sites{
		newJobs, err := o.scraper.ScrapeAndStoreJobs(ctx, site)
		if err != nil {
			log.Printf("error to scraping site %s : %v", site.SiteName, err)
		}

		if len(newJobs) == 0{
			continue
		}

		if err := o.notifier.FindMatchesAndNotify(site.ID, newJobs); err != nil {
			log.Printf("error to process notification for site: %s : %v", site.SiteName, err)
		}
	}
}