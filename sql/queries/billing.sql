-- name: ListPlans :many
SELECT *
FROM plans
WHERE is_active = true
ORDER BY sort_order ASC, price_cents ASC;

-- name: ListPlansAdmin :many
SELECT *
FROM plans
ORDER BY sort_order ASC, price_cents ASC;

-- name: GetPlan :one
SELECT *
FROM plans
WHERE slug = $1;

-- name: GetActivePlan :one
SELECT *
FROM plans
WHERE slug = $1 AND is_active = true;

-- name: GetDefaultPlan :one
SELECT *
FROM plans
WHERE is_active = true
ORDER BY sort_order ASC, price_cents ASC
LIMIT 1;

-- name: CreatePlan :one
INSERT INTO plans (
    slug,
    name,
    price_cents,
    stripe_price_id,
    messages_limit,
    connections_limit,
    retention_days,
    webhooks_enabled,
    video_calls_enabled,
    moderation_enabled,
    is_active,
    is_highlighted,
    sort_order,
    description
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: UpdatePlan :one
UPDATE plans
SET
    name = COALESCE(sqlc.narg('name'), name),
    price_cents = COALESCE(sqlc.narg('price_cents'), price_cents),
    stripe_price_id = COALESCE(sqlc.narg('stripe_price_id'), stripe_price_id),
    messages_limit = COALESCE(sqlc.narg('messages_limit'), messages_limit),
    connections_limit = COALESCE(sqlc.narg('connections_limit'), connections_limit),
    retention_days = COALESCE(sqlc.narg('retention_days'), retention_days),
    webhooks_enabled = COALESCE(sqlc.narg('webhooks_enabled'), webhooks_enabled),
    video_calls_enabled = COALESCE(sqlc.narg('video_calls_enabled'), video_calls_enabled),
    moderation_enabled = COALESCE(sqlc.narg('moderation_enabled'), moderation_enabled),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    is_highlighted = COALESCE(sqlc.narg('is_highlighted'), is_highlighted),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = NOW()
WHERE slug = sqlc.arg('slug')
RETURNING *;

-- name: DeletePlan :exec
DELETE FROM plans
WHERE slug = $1;

-- name: CountSubscriptionsForPlan :one
SELECT COUNT(*)::bigint
FROM app_subscriptions
WHERE plan_slug = $1;

-- name: GetAppByWorkspaceID :one
SELECT *
FROM apps
WHERE workspace_id = $1
   OR sandbox_workspace_id = $1;

-- name: CreateAppSubscription :one
INSERT INTO app_subscriptions (
    app_id,
    plan_slug,
    status
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAppSubscription :one
SELECT
    s.*,
    p.name AS plan_name,
    p.price_cents,
    p.messages_limit,
    p.connections_limit,
    p.retention_days,
    p.webhooks_enabled,
    p.video_calls_enabled,
    p.moderation_enabled
FROM app_subscriptions s
INNER JOIN plans p ON p.slug = s.plan_slug
WHERE s.app_id = $1;

-- name: UpdateAppSubscription :one
UPDATE app_subscriptions
SET
    plan_slug = COALESCE(sqlc.narg('plan_slug'), plan_slug),
    status = COALESCE(sqlc.narg('status'), status),
    stripe_customer_id = COALESCE(sqlc.narg('stripe_customer_id'), stripe_customer_id),
    stripe_subscription_id = COALESCE(sqlc.narg('stripe_subscription_id'), stripe_subscription_id),
    current_period_start = COALESCE(sqlc.narg('current_period_start'), current_period_start),
    current_period_end = COALESCE(sqlc.narg('current_period_end'), current_period_end),
    updated_at = NOW()
WHERE app_id = sqlc.arg('app_id')
RETURNING *;

-- name: GetAppSubscriptionByStripeSubscriptionID :one
SELECT *
FROM app_subscriptions
WHERE stripe_subscription_id = $1;

-- name: IncrementAppMessageUsage :one
INSERT INTO usage_counters (app_id, period_month, messages_count)
VALUES ($1, date_trunc('month', NOW())::date, 1)
ON CONFLICT (app_id, period_month)
DO UPDATE SET
    messages_count = usage_counters.messages_count + 1,
    updated_at = NOW()
RETURNING messages_count;

-- name: GetAppMessageUsage :one
SELECT messages_count
FROM usage_counters
WHERE app_id = $1
  AND period_month = date_trunc('month', NOW())::date;

-- name: SetUserStripeCustomerID :exec
UPDATE users
SET stripe_customer_id = $2
WHERE id = $1
  AND (stripe_customer_id IS NULL OR stripe_customer_id = '');

-- name: GetUserStripeCustomerID :one
SELECT stripe_customer_id
FROM users
WHERE id = $1;

-- name: GetPlatformSetting :one
SELECT value
FROM platform_settings
WHERE key = $1;

-- name: UpsertPlatformSetting :exec
INSERT INTO platform_settings (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = NOW();

-- name: AdminCountActiveSubscriptions :one
SELECT COUNT(*)::bigint
FROM app_subscriptions
WHERE status IN ('active', 'trialing');

-- name: AdminListAppSubscriptions :many
SELECT
    s.app_id,
    a.name AS app_name,
    a.slug AS app_slug,
    s.plan_slug,
    p.name AS plan_name,
    s.status,
    s.current_period_end,
    COALESCE(u.messages_count, 0)::bigint AS messages_this_month
FROM app_subscriptions s
INNER JOIN apps a ON a.id = s.app_id
INNER JOIN plans p ON p.slug = s.plan_slug
LEFT JOIN usage_counters u
    ON u.app_id = s.app_id
   AND u.period_month = date_trunc('month', NOW())::date
ORDER BY s.updated_at DESC
LIMIT $1 OFFSET $2;
