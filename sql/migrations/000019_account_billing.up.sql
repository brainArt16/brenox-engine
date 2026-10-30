-- Account-level subscriptions: one plan per user, shared message pool, max_apps.

ALTER TABLE plans
    ADD COLUMN IF NOT EXISTS max_apps INT NOT NULL DEFAULT 1;

UPDATE plans SET max_apps = 2, price_cents = 0 WHERE slug = 'starter';
UPDATE plans SET max_apps = 5 WHERE slug = 'growth';
UPDATE plans SET max_apps = 20 WHERE slug = 'scale';

CREATE TABLE user_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    plan_slug VARCHAR(32) NOT NULL REFERENCES plans(slug),
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    stripe_customer_id VARCHAR(128),
    stripe_subscription_id VARCHAR(128),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_subscriptions_plan ON user_subscriptions(plan_slug);
CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);

-- Free Starter for every existing user.
INSERT INTO user_subscriptions (user_id, plan_slug, status)
SELECT id, 'starter', 'active'
FROM users
ON CONFLICT (user_id) DO NOTHING;

-- Prefer highest-priced app subscription (paid upgrades) when present.
WITH best AS (
    SELECT DISTINCT ON (a.owner_id)
        a.owner_id AS user_id,
        s.plan_slug,
        s.status,
        s.stripe_customer_id,
        s.stripe_subscription_id,
        s.current_period_start,
        s.current_period_end,
        p.price_cents
    FROM app_subscriptions s
    INNER JOIN apps a ON a.id = s.app_id
    INNER JOIN plans p ON p.slug = s.plan_slug
    ORDER BY
        a.owner_id,
        p.price_cents DESC,
        CASE s.status
            WHEN 'active' THEN 0
            WHEN 'trialing' THEN 1
            WHEN 'past_due' THEN 2
            ELSE 3
        END,
        s.updated_at DESC
)
UPDATE user_subscriptions us
SET
    plan_slug = best.plan_slug,
    status = CASE
        WHEN best.price_cents = 0 THEN 'active'
        WHEN best.status IN ('incomplete', 'incomplete_expired') THEN 'active'
        ELSE best.status
    END,
    stripe_customer_id = COALESCE(best.stripe_customer_id, us.stripe_customer_id),
    stripe_subscription_id = COALESCE(best.stripe_subscription_id, us.stripe_subscription_id),
    current_period_start = COALESCE(best.current_period_start, us.current_period_start),
    current_period_end = COALESCE(best.current_period_end, us.current_period_end),
    updated_at = NOW()
FROM best
WHERE us.user_id = best.user_id
  AND (best.price_cents > 0 OR best.stripe_subscription_id IS NOT NULL);

DROP TABLE IF EXISTS app_subscriptions;
