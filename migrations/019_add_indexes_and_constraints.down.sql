-- Revert user_sites.filters back to JSON
ALTER TABLE user_sites ALTER COLUMN filters TYPE JSON USING filters::JSON;

-- Revert requested_sites foreign key (remove CASCADE)
ALTER TABLE requested_sites
    DROP CONSTRAINT IF EXISTS requested_sites_user_id_fkey,
    ADD CONSTRAINT requested_sites_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

-- Revert user_sites foreign keys (remove CASCADE)
ALTER TABLE user_sites
    DROP CONSTRAINT IF EXISTS user_sites_site_id_fkey,
    ADD CONSTRAINT user_sites_site_id_fkey FOREIGN KEY (site_id) REFERENCES site_scraping_config(id);

ALTER TABLE user_sites
    DROP CONSTRAINT IF EXISTS user_sites_user_id_fkey,
    ADD CONSTRAINT user_sites_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

-- Drop indexes
DROP INDEX IF EXISTS idx_job_notifications_sent_at;
DROP INDEX IF EXISTS idx_curriculum_user_id;
DROP INDEX IF EXISTS idx_jobs_company;
DROP INDEX IF EXISTS idx_jobs_requisition_id;
