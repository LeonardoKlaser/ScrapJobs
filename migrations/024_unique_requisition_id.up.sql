-- Remove duplicates keeping the latest (by id)
DELETE FROM jobs a USING jobs b
WHERE a.id < b.id AND a.requisition_id = b.requisition_id;

ALTER TABLE jobs ADD CONSTRAINT uq_jobs_requisition_id UNIQUE (requisition_id);
