ALTER TABLE jobs
    DROP COLUMN IF EXISTS job_link,
    DROP COLUMN IF EXISTS requisition_id;
