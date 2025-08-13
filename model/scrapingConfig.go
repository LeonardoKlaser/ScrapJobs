// model/scrapingConfig.go
package model


type SiteScrapingConfig struct {
	ID        int    `db:"id"`
	SiteName  string `db:"site_name"`
	BaseURL   string `db:"base_url"`
	IsActive  bool   `db:"is_active"`
	ScrapingType string `db:"scraping_type"` // 'HTML', 'API', 'HEADLESS'
	JobListItemSelector      *string `db:"job_list_item_selector"`
	TitleSelector            *string `db:"title_selector"`
	LinkSelector             *string `db:"link_selector"`
	LinkAttribute            *string `db:"link_attribute"`
	LocationSelector         *string `db:"location_selector"`
	NextPageSelector         *string `db:"next_page_selector"`
	JobDescriptionSelector   *string `db:"job_description_selector"`
	JobRequisitionIdSelector *string `db:"job_requisition_id_selector"`
	APIEndpointTemplate *string `db:"api_endpoint_template"`
	APIMethod           *string `db:"api_method"`
	APIHeadersJSON      *string `db:"api_headers_json"` // JSON como string
	APIPayloadTemplate  *string `db:"api_payload_template"`
	JSONDataMappings    *string `db:"json_data_mappings"` // JSON como string
}