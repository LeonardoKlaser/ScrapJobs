-- Drop the old single-column index on sent_at (now notified_at)
DROP INDEX IF EXISTS idx_job_notifications_sent_at;

-- Create composite index for monthly count query
CREATE INDEX IF NOT EXISTS idx_job_notifications_user_notified ON job_notifications(user_id, notified_at);
