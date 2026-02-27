ALTER TABLE job_notifications DROP COLUMN IF EXISTS curriculum_id;
ALTER TABLE job_notifications DROP COLUMN IF EXISTS analysis_result;
ALTER TABLE curriculum ADD COLUMN is_active BOOLEAN DEFAULT FALSE;
CREATE UNIQUE INDEX partial_unique_active_curriculum ON curriculum (user_id) WHERE is_active = TRUE;
