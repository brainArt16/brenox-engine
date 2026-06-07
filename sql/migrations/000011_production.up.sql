CREATE TABLE revoked_tokens (
    jti TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_revoked_tokens_user_id ON revoked_tokens (user_id);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    app_id BIGINT REFERENCES apps(id),
    action TEXT NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    ip_address TEXT,
    status_code INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);
