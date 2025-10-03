package model

type SiteScrapingConfig struct {
	ID                       int     `db:"id" json:"id"`
	SiteName                 string  `db:"site_name" json:"site_name"`
	BaseURL                  string  `db:"base_url" json:"base_url"`
	LogoURL                  *string  `db:"logo_url" json:"logo_url,omitempty"`
	IsActive                 bool    `db:"is_active" json:"is_active"`
	ScrapingType             string  `db:"scraping_type" json:"scraping_type"` // 'HTML', 'API', 'HEADLESS'
	JobListItemSelector      *string `db:"job_list_item_selector" json:"job_list_item_selector,omitempty"`
	TitleSelector            *string `db:"title_selector" json:"title_selector,omitempty"`
	LinkSelector             *string `db:"link_selector" json:"link_selector,omitempty"`
	LinkAttribute            *string `db:"link_attribute" json:"link_attribute,omitempty"`
	LocationSelector         *string `db:"location_selector" json:"location_selector,omitempty"`
	NextPageSelector         *string `db:"next_page_selector" json:"next_page_selector,omitempty"`
	JobDescriptionSelector   *string `db:"job_description_selector" json:"job_description_selector,omitempty"`
	JobRequisitionIdSelector *string `db:"job_requisition_id_selector" json:"job_requisition_id_selector,omitempty"`
	APIEndpointTemplate      *string `db:"api_endpoint_template" json:"api_endpoint_template,omitempty"`
	APIMethod                *string `db:"api_method" json:"api_method,omitempty"`
	APIHeadersJSON           *string `db:"api_headers_json" json:"api_headers_json,omitempty"`
	APIPayloadTemplate       *string `db:"api_payload_template" json:"api_payload_template,omitempty"`
	JSONDataMappings         *string `db:"json_data_mappings" json:"json_data_mappings,omitempty"`
}