-- name: CreateApp :one
INSERT INTO apps (
    name,
    slug,
    workspace_id,
    owner_id
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetAppByID :one
SELECT *
FROM apps
WHERE id = $1;

-- name: GetAppBySlug :one
SELECT *
FROM apps
WHERE slug = $1;

-- name: ListAppsByOwner :many
SELECT *
FROM apps
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO api_keys (
    app_id,
    name,
    key_prefix,
    key_hash,
    is_sandbox
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetAPIKeyByPrefix :one
SELECT *
FROM api_keys
WHERE key_prefix = $1
  AND revoked_at IS NULL;

-- name: ListAPIKeysByApp :many
SELECT *
FROM api_keys
WHERE app_id = $1
ORDER BY created_at DESC;

-- name: RevokeAPIKey :one
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1
  AND app_id = $2
  AND revoked_at IS NULL
RETURNING *;

-- name: TouchAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;

-- name: CreateAppUser :one
INSERT INTO app_users (
    app_id,
    user_id,
    external_id
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetAppUserByExternalID :one
SELECT *
FROM app_users
WHERE app_id = $1
  AND external_id = $2;

-- name: GetAppUserByUserID :one
SELECT *
FROM app_users
WHERE app_id = $1
  AND user_id = $2;

-- name: CreateWebhook :one
INSERT INTO webhooks (
    app_id,
    url,
    events,
    secret
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: ListWebhooksByApp :many
SELECT *
FROM webhooks
WHERE app_id = $1
  AND disabled_at IS NULL
ORDER BY created_at DESC;

-- name: GetIdempotencyKey :one
SELECT *
FROM idempotency_keys
WHERE app_id = $1
  AND idempotency_key = $2;

-- name: CreateIdempotencyKey :one
INSERT INTO idempotency_keys (
    app_id,
    idempotency_key,
    endpoint,
    status_code,
    response_body
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
