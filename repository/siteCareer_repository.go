package repository

import (
	"database/sql"
	"web-scrapper/model"
	"encoding/json"
	"fmt"
)

type SiteCareerRepository struct {
	connection *sql.DB
}

func NewSiteCareerRepository(db *sql.DB) SiteCareerRepository{
	return SiteCareerRepository{
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
	var targetWordsJSON []byte

	targetWords, err := json.Marshal(site.TargetWords)
	if err != nil {
		return nilReturn, fmt.Errorf("error to serialize target words: %w", err)
	}

	err = queryPrepare.QueryRow(site.SiteName, site.BaseURL, site.JobListItemSelector, site.TitleSelector, site.LinkSelector, site.LinkAttribute,
								site.LocationSelector, site.NextPageSelector, site.JobDescriptionSelector, site.JobRequisitionIdSelector, targetWords).Scan(
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
										targetWordsJSON,
								)
	if err != nil {
		if err == sql.ErrNoRows{
			return nilReturn, fmt.Errorf("error to insert data in the database: %w", err)
		}
		return nilReturn, err
	}
	
	if len(targetWordsJSON) > 0 {
		if err := json.Unmarshal(targetWordsJSON, &siteCreated.TargetWords); err != nil {
			return nilReturn, fmt.Errorf("error to get target words informations: %w", err )
		}
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

	var targetWordsJSON []byte

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
			&targetWordsJSON,
		)

		if err != nil{
			if err == sql.ErrNoRows{
				return []model.SiteScrapingConfig{}, fmt.Errorf("error to get site: %w", err)
			}
			return []model.SiteScrapingConfig{}, err
		}

		if len(targetWordsJSON) > 0 {
			if err := json.Unmarshal(targetWordsJSON, &site.TargetWords); err != nil{
				return []model.SiteScrapingConfig{}, fmt.Errorf("error to get targetWords: %w", err)
			}
		}

		listOfSites = append(listOfSites, site)
	}

	return listOfSites, nil
}