ALTER TABLE plans ADD CONSTRAINT chk_max_ai_analyses CHECK (max_ai_analyses >= -1);
ALTER TABLE requested_sites ADD CONSTRAINT chk_requested_sites_status CHECK (status IN ('pending', 'approved', 'rejected'));
