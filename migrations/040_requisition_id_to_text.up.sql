ALTER TABLE jobs DROP CONSTRAINT IF EXISTS uq_jobs_requisition_id;
ALTER TABLE jobs ALTER COLUMN requisition_id TYPE TEXT USING requisition_id::TEXT;
CREATE UNIQUE INDEX uq_jobs_requisition_id ON jobs (requisition_id) WHERE requisition_id IS NOT NULL AND requisition_id != '';
