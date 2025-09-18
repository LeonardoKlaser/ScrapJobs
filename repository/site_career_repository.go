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

	query := `
        INSERT INTO site_scraping_config (
            site_name, base_url, is_active, scraping_type,
            job_list_item_selector, title_selector, link_selector, link_attribute,
            location_selector, next_page_selector, job_description_selector, job_requisition_id_selector,
            api_endpoint_template, api_method, api_headers_json, api_payload_template, json_data_mappings
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
        ) RETURNING 
            id, site_name, base_url, is_active, scraping_type,
            job_list_item_selector, title_selector, link_selector, link_attribute,
            location_selector, next_page_selector, job_description_selector, job_requisition_id_selector,
            api_endpoint_template, api_method, api_headers_json, api_payload_template, json_data_mappings
    `
	var siteCreated model.SiteScrapingConfig


	err := st.connection.QueryRow(
		query,
		site.SiteName, site.BaseURL, site.IsActive, site.ScrapingType,
		site.JobListItemSelector, site.TitleSelector, site.LinkSelector, site.LinkAttribute,
		site.LocationSelector, site.NextPageSelector, site.JobDescriptionSelector, site.JobRequisitionIdSelector,
		site.APIEndpointTemplate, site.APIMethod, site.APIHeadersJSON, site.APIPayloadTemplate, site.JSONDataMappings,
	).Scan(
		&siteCreated.ID, &siteCreated.SiteName, &siteCreated.BaseURL, &siteCreated.IsActive, &siteCreated.ScrapingType,
		&siteCreated.JobListItemSelector, &siteCreated.TitleSelector, &siteCreated.LinkSelector, &siteCreated.LinkAttribute,
		&siteCreated.LocationSelector, &siteCreated.NextPageSelector, &siteCreated.JobDescriptionSelector, &siteCreated.JobRequisitionIdSelector,
		&siteCreated.APIEndpointTemplate, &siteCreated.APIMethod, &siteCreated.APIHeadersJSON, &siteCreated.APIPayloadTemplate, &siteCreated.JSONDataMappings,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nilReturn, fmt.Errorf("erro ao inserir dados no banco de dados: %w", err)
		}
		return nilReturn, err
	}

	return siteCreated, nil
	
}

func (st *SiteCareerRepository) GetAllSites() ([]model.SiteScrapingConfig, error){
	var listOfSites []model.SiteScrapingConfig;

	query := "SELECT * FROM site_scraping_config WHERE is_active = TRUE"
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
			&site.IsActive,
			&site.ScrapingType,
			&site.JobListItemSelector,
			&site.TitleSelector,
			&site.LinkSelector,
			&site.LinkAttribute,
			&site.LocationSelector,
			&site.NextPageSelector,
			&site.JobDescriptionSelector,
			&site.JobRequisitionIdSelector,
			&site.APIEndpointTemplate,
			&site.APIMethod,
			&site.APIHeadersJSON,
			&site.APIPayloadTemplate,
			&site.JSONDataMappings,
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