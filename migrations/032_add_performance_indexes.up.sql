-- user_sites.user_id — usado em GetSubscribedSiteIDs, GetUserSiteCount, DeleteUserSite, dashboard joins
CREATE INDEX IF NOT EXISTS idx_user_sites_user_id ON user_sites(user_id);

-- jobs.created_at — usado em dashboard cards e GetLatestJobsPaginated
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);

-- jobs.last_seen_at — usado em DeleteOldJobs e GetUnnotifiedJobsForUser
CREATE INDEX IF NOT EXISTS idx_jobs_last_seen_at ON jobs(last_seen_at DESC);

-- job_notifications.job_id — usado em GetNotifiedJobIDsForUser e BulkInsert ON CONFLICT
CREATE INDEX IF NOT EXISTS idx_job_notifications_job_id ON job_notifications(job_id);

-- site_scraping_config ativa — usado em GetAllSites WHERE is_active = TRUE
CREATE INDEX IF NOT EXISTS idx_site_config_active ON site_scraping_config(id) WHERE is_active = TRUE;
