CREATE TABLE IF NOT EXISTS scraping_errors (
    id SERIAL PRIMARY KEY,
    site_id INTEGER REFERENCES site_scraping_config(id) ON DELETE SET NULL,
    site_name VARCHAR(255) NOT NULL,
    error_message TEXT NOT NULL,
    task_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scraping_errors_created_at ON scraping_errors(created_at);
