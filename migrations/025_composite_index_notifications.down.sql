DROP INDEX IF EXISTS idx_job_notifications_user_notified;
CREATE INDEX IF NOT EXISTS idx_job_notifications_sent_at ON job_notifications(notified_at);
