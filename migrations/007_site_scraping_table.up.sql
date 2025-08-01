CREATE TYPE scraping_strategy AS ENUM ('HTML', 'API', 'HEADLESS');

CREATE TABLE site_scraping_config (
    id SERIAL PRIMARY KEY,
    site_name VARCHAR(100) NOT NULL,
    base_url TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,

    scraping_type scraping_strategy NOT NULL,

    job_list_item_selector TEXT,
    title_selector TEXT,
    link_selector TEXT,
    link_attribute VARCHAR(50),
    location_selector TEXT,
    next_page_selector TEXT,
    job_description_selector TEXT,
    job_requisition_id_selector TEXT,

    
    api_endpoint_template TEXT,
    api_method VARCHAR(10) DEFAULT 'GET',
    api_headers_json JSONB,
    api_payload_template TEXT,
    json_data_mappings JSONB NOT NULL
);