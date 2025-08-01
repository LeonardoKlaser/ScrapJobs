// model/scrapingConfig.go
package model

import "database/sql"

// SiteScrapingConfig reflete a nova tabela do banco de dados.
type SiteScrapingConfig struct {
	ID        int    `db:"id"`
	SiteName  string `db:"site_name"`
	BaseURL   string `db:"base_url"`
	IsActive  bool   `db:"is_active"`
	ScrapingType string `db:"scraping_type"` // 'HTML', 'API', 'HEADLESS'
	JobListItemSelector      sql.NullString `db:"job_list_item_selector"`
	TitleSelector            sql.NullString `db:"title_selector"`
	LinkSelector             sql.NullString `db:"link_selector"`
	LinkAttribute            sql.NullString `db:"link_attribute"`
	LocationSelector         sql.NullString `db:"location_selector"`
	NextPageSelector         sql.NullString `db:"next_page_selector"`
	JobDescriptionSelector   sql.NullString `db:"job_description_selector"`
	JobRequisitionIdSelector sql.NullString `db:"job_requisition_id_selector"`
	APIEndpointTemplate sql.NullString `db:"api_endpoint_template"`
	APIMethod           sql.NullString `db:"api_method"`
	APIHeadersJSON      sql.NullString `db:"api_headers_json"` // JSON como string
	APIPayloadTemplate  sql.NullString `db:"api_payload_template"`
	JSONDataMappings    sql.NullString `db:"json_data_mappings"` // JSON como string
}