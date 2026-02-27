ALTER TABLE curriculum DROP CONSTRAINT IF EXISTS partial_unique_active_curriculum;
ALTER TABLE curriculum DROP COLUMN IF EXISTS is_active;
ALTER TABLE job_notifications ADD COLUMN analysis_result JSONB NULL;
ALTER TABLE job_notifications ADD COLUMN curriculum_id INT NULL;
