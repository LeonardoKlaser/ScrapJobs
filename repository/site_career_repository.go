package repository

import (
	"database/sql"
	"web-scrapper/model"
	"fmt"
)

type SiteCareerRepository struct {
	connection *sql.DB
}

func NewSiteCareerRepository(db *sql.DB) *SiteCareerRepository{
	return &SiteCareerRepository{
		connection: db,
	}
}

func (st *SiteCareerRepository) InsertNewSiteCareer(site model.SiteScrapingConfig) (model.SiteScrapingConfig, error){
	nilReturn := model.SiteScrapingConfig{}

	query := "INSERT INTO site_scraping_config (site_name, base_url, job_list_item_selector, title_selector, link_selector, link_attribute," +
									" location_selector, next_page_selector, job_descrpt_selector, job_req_id_selector, target_words) " +
									"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *"
	queryPrepare, err := st.connection.Prepare(query)

	if err != nil {
		return nilReturn, fmt.Errorf("error to run query in the database: %w", err)
	}

	var siteCreated model.SiteScrapingConfig


	err = queryPrepare.QueryRow(site.SiteName, site.BaseURL, site.JobListItemSelector, site.TitleSelector, site.LinkSelector, site.LinkAttribute,
								site.LocationSelector, site.NextPageSelector, site.JobDescriptionSelector, site.JobRequisitionIdSelector).Scan(
										&siteCreated.SiteName,
										&siteCreated.BaseURL,
										&siteCreated.JobListItemSelector,
										&siteCreated.TitleSelector,
										&siteCreated.LinkSelector,
										&siteCreated.LinkAttribute,
										&siteCreated.LocationSelector,
										&siteCreated.NextPageSelector,
										&siteCreated.JobDescriptionSelector,
										&siteCreated.JobRequisitionIdSelector,
								)
	if err != nil {
		if err == sql.ErrNoRows{
			return nilReturn, fmt.Errorf("error to insert data in the database : %w", err)
		}
		return nilReturn, err
	}
	

	queryPrepare.Close()
	return siteCreated, nil
	
}

func (st *SiteCareerRepository) GetAllSites() ([]model.SiteScrapingConfig, error){
	var listOfSites []model.SiteScrapingConfig;

	query := "SELECT * FROM site_scraping_config"
	rows, err := st.connection.Query(query)

	if err != nil {
		return listOfSites, fmt.Errorf("error to querie: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var site model.SiteScrapingConfig
		err := rows.Scan(
			&site.ID,
			&site.SiteName,
			&site.BaseURL,
			&site.JobListItemSelector,
			&site.TitleSelector,
			&site.LinkSelector,
			&site.LinkAttribute,
			&site.LocationSelector,
			&site.NextPageSelector,
			&site.JobDescriptionSelector,
			&site.JobRequisitionIdSelector,
		)

		if err != nil{
			if err == sql.ErrNoRows{
				return []model.SiteScrapingConfig{}, fmt.Errorf("error to get site: %w", err)
			}
			return []model.SiteScrapingConfig{}, err
		}


		listOfSites = append(listOfSites, site)
	}

	return listOfSites, nil
}