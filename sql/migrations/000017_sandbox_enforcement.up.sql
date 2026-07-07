ALTER TABLE api_keys
    ADD COLUMN expires_at TIMESTAMPTZ;

UPDATE api_keys
SET expires_at = NOW() + INTERVAL '90 days'
WHERE is_sandbox = true
  AND revoked_at IS NULL
  AND expires_at IS NULL;

CREATE INDEX idx_api_keys_expires_at
    ON api_keys (expires_at)
    WHERE expires_at IS NOT NULL;

CREATE INDEX idx_app_users_app_environment
    ON app_users (app_id, environment);

CREATE INDEX idx_messages_created_at
    ON messages (created_at);
