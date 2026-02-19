ALTER TABLE job_notifications DROP CONSTRAINT IF EXISTS uq_job_notifications_user_job;
ALTER TABLE job_notifications RENAME COLUMN notified_at TO sent_at;
ALTER TABLE job_notifications DROP CONSTRAINT IF EXISTS job_notifications_pkey;
ALTER TABLE job_notifications DROP COLUMN IF EXISTS id;
ALTER TABLE job_notifications ADD PRIMARY KEY (user_id, job_id);
