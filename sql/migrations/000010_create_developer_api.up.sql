CREATE TABLE apps (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    workspace_id BIGINT NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE,
    owner_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT 'default',
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    is_sandbox BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_api_keys_prefix_active
    ON api_keys (key_prefix)
    WHERE revoked_at IS NULL;

CREATE TABLE app_users (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, external_id),
    UNIQUE (app_id, user_id)
);

CREATE TABLE webhooks (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    secret TEXT NOT NULL,
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    status_code INT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, idempotency_key)
);

CREATE INDEX idx_app_users_app_id ON app_users (app_id);
CREATE INDEX idx_webhooks_app_id ON webhooks (app_id);
