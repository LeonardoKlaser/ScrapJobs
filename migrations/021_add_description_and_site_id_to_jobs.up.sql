ALTER TABLE jobs ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS site_id INTEGER REFERENCES site_scraping_config(id);
CREATE INDEX IF NOT EXISTS idx_jobs_site_id ON jobs(site_id);
