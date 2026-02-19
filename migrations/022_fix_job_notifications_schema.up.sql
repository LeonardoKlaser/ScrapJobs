-- Add id column as serial primary key
ALTER TABLE job_notifications DROP CONSTRAINT IF EXISTS job_notifications_pkey;
ALTER TABLE job_notifications ADD COLUMN IF NOT EXISTS id SERIAL;
ALTER TABLE job_notifications ADD PRIMARY KEY (id);

-- Rename sent_at to notified_at
ALTER TABLE job_notifications RENAME COLUMN sent_at TO notified_at;

-- Add unique constraint for dedup
ALTER TABLE job_notifications ADD CONSTRAINT uq_job_notifications_user_job UNIQUE (user_id, job_id);
