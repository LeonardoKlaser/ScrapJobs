package model


type SiteScrapingConfig struct {
	ID 			int `db:"id"`
	SiteName    string `db:"site_name"`
	BaseURL     string `db:"base_url"`
	JobListItemSelector string `db:"job_list_item_selector"`
	TitleSelector string `db:"title_selector"`
	LinkSelector  string `db:"link_selector"`
	LinkAttribute string `db:"link_attribute"`
	LocationSelector string `db:"location_selector"`
	NextPageSelector string `db:"next_page_selector"`
	JobDescriptionSelector string `db:"job_descrpt_selector"`
	JobRequisitionIdSelector string `db:"job_req_id_selector"`
}