DROP INDEX IF EXISTS idx_job_notifications_status;
ALTER TABLE job_notifications DROP COLUMN IF EXISTS status;
