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

-- Restore one app subscription per app from the owner's account plan.
INSERT INTO app_subscriptions (app_id, plan_slug, status, stripe_customer_id, stripe_subscription_id, current_period_start, current_period_end)
SELECT
    a.id,
    us.plan_slug,
    us.status,
    us.stripe_customer_id,
    us.stripe_subscription_id,
    us.current_period_start,
    us.current_period_end
FROM apps a
INNER JOIN user_subscriptions us ON us.user_id = a.owner_id
ON CONFLICT (app_id) DO NOTHING;

DROP TABLE IF EXISTS user_subscriptions;

ALTER TABLE plans DROP COLUMN IF EXISTS max_apps;

UPDATE plans SET price_cents = 1000 WHERE slug = 'starter' AND price_cents = 0;
