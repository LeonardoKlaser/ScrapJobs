-- Ensure at most one active curriculum per user at the DB level
CREATE UNIQUE INDEX IF NOT EXISTS uq_curriculum_active_per_user ON curriculum(user_id) WHERE is_active = TRUE;
