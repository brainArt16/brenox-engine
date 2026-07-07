ALTER TABLE users
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(128);

CREATE TABLE plans (
    slug VARCHAR(32) PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    price_cents INT NOT NULL,
    stripe_price_id VARCHAR(128),
    messages_limit INT NOT NULL,
    connections_limit INT NOT NULL,
    retention_days INT NOT NULL,
    webhooks_enabled BOOLEAN NOT NULL DEFAULT false,
    video_calls_enabled BOOLEAN NOT NULL DEFAULT false,
    moderation_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO plans (
    slug,
    name,
    price_cents,
    messages_limit,
    connections_limit,
    retention_days,
    webhooks_enabled,
    video_calls_enabled,
    moderation_enabled
) VALUES
    ('starter', 'Starter', 1000, 50000, 500, 30, false, false, false),
    ('growth', 'Growth', 2500, 250000, 2500, 90, true, true, false),
    ('scale', 'Scale', 5000, 1000000, 10000, 365, true, true, true);

CREATE TABLE app_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL UNIQUE REFERENCES apps(id) ON DELETE CASCADE,
    plan_slug VARCHAR(32) NOT NULL REFERENCES plans(slug),
    status VARCHAR(32) NOT NULL DEFAULT 'incomplete',
    stripe_customer_id VARCHAR(128),
    stripe_subscription_id VARCHAR(128),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_subscriptions_plan ON app_subscriptions(plan_slug);
CREATE INDEX idx_app_subscriptions_status ON app_subscriptions(status);

INSERT INTO app_subscriptions (app_id, plan_slug, status)
SELECT id, 'starter', 'incomplete'
FROM apps
ON CONFLICT (app_id) DO NOTHING;

CREATE TABLE usage_counters (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    period_month DATE NOT NULL,
    messages_count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, period_month)
);

CREATE TABLE platform_settings (
    key VARCHAR(64) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO platform_settings (key, value) VALUES
    ('maintenance_mode', 'false'),
    ('maintenance_message', 'Brenox is undergoing scheduled maintenance. Please try again shortly.');
