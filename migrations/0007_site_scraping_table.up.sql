CREATE TABLE site_scraping_config (
    id SERIAL PRIMARY KEY,
    site_name VARCHAR(100) NOT NULL,
    base_url TEXT NOT NULL,
    job_list_item_selector TEXT NOT NULL,
    title_selector TEXT NOT NULL,
    link_selector TEXT NOT NULL,
    link_attribute VARCHAR(50) NOT NULL,
    location_selector TEXT NOT NULL,
    next_page_selector TEXT,
    job_description_selector TEXT,
    job_requisition_id_selector TEXT,
    target_words JSON 
);