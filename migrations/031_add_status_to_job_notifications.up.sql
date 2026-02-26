ALTER TABLE job_notifications
ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'SENT';

CREATE INDEX idx_job_notifications_status
ON job_notifications(user_id, status)
WHERE status = 'PENDING';
