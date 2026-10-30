-- Per-plan video limits: monthly minutes (account pool) + max participants.

ALTER TABLE plans
    ADD COLUMN IF NOT EXISTS video_minutes_limit INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS max_call_participants INT NOT NULL DEFAULT 0;

UPDATE plans SET video_minutes_limit = 0, max_call_participants = 5 WHERE slug = 'starter';
UPDATE plans SET video_minutes_limit = 1000, max_call_participants = 15 WHERE slug = 'growth';
UPDATE plans SET video_minutes_limit = 10000, max_call_participants = 25 WHERE slug = 'scale';

ALTER TABLE usage_counters
    ADD COLUMN IF NOT EXISTS video_minutes BIGINT NOT NULL DEFAULT 0;
