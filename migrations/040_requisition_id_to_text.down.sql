DROP INDEX IF EXISTS uq_jobs_requisition_id;
ALTER TABLE jobs ALTER COLUMN requisition_id TYPE BIGINT USING requisition_id::BIGINT;
ALTER TABLE jobs ADD CONSTRAINT uq_jobs_requisition_id UNIQUE (requisition_id);
