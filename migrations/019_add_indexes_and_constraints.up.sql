-- Indexes for frequently queried columns
CREATE INDEX IF NOT EXISTS idx_jobs_requisition_id ON jobs(requisition_id);
CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company);
CREATE INDEX IF NOT EXISTS idx_curriculum_user_id ON curriculum(user_id);
CREATE INDEX IF NOT EXISTS idx_job_notifications_sent_at ON job_notifications(sent_at);

-- Add ON DELETE CASCADE to user_sites foreign keys
ALTER TABLE user_sites
    DROP CONSTRAINT IF EXISTS user_sites_user_id_fkey,
    ADD CONSTRAINT user_sites_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_sites
    DROP CONSTRAINT IF EXISTS user_sites_site_id_fkey,
    ADD CONSTRAINT user_sites_site_id_fkey FOREIGN KEY (site_id) REFERENCES site_scraping_config(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to requested_sites foreign key
ALTER TABLE requested_sites
    DROP CONSTRAINT IF EXISTS requested_sites_user_id_fkey,
    ADD CONSTRAINT requested_sites_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Change user_sites.filters from JSON to JSONB for better query performance
ALTER TABLE user_sites ALTER COLUMN filters TYPE JSONB USING filters::JSONB;
