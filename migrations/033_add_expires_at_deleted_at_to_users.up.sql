ALTER TABLE users ADD COLUMN expires_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;
UPDATE users SET expires_at = NOW() + INTERVAL '30 days' WHERE plan_id IS NOT NULL;
CREATE INDEX idx_users_expires_at ON users (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;
