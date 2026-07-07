-- name: AdminCountUsers :one
SELECT COUNT(*)::bigint AS count FROM users;

-- name: AdminCountWorkspaces :one
SELECT COUNT(*)::bigint AS count FROM workspaces;

-- name: AdminCountApps :one
SELECT COUNT(*)::bigint AS count FROM apps;

-- name: ListUsersAdmin :many
SELECT
    id,
    email,
    username,
    platform_role,
    suspended_at,
    created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SearchUsersAdmin :many
SELECT
    id,
    email,
    username,
    platform_role,
    suspended_at,
    created_at
FROM users
WHERE email ILIKE '%' || @search::text || '%'
   OR username ILIKE '%' || @search::text || '%'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetUserAdmin :one
SELECT
    id,
    email,
    username,
    platform_role,
    suspended_at,
    tokens_invalidated_at,
    created_at
FROM users
WHERE id = $1;

-- name: CountUserWorkspaces :one
SELECT COUNT(*)::bigint AS count
FROM workspace_members
WHERE user_id = $1;

-- name: CountUserApps :one
SELECT COUNT(*)::bigint AS count
FROM apps
WHERE owner_id = $1;

-- name: UpdateUserPlatformRole :one
UPDATE users
SET platform_role = $2
WHERE id = $1
RETURNING id, email, username, platform_role, suspended_at, tokens_invalidated_at, created_at;

-- name: SuspendUser :one
UPDATE users
SET
    suspended_at = NOW(),
    tokens_invalidated_at = NOW()
WHERE id = $1
RETURNING id, email, username, platform_role, suspended_at, tokens_invalidated_at, created_at;

-- name: UnsuspendUser :one
UPDATE users
SET suspended_at = NULL
WHERE id = $1
RETURNING id, email, username, platform_role, suspended_at, tokens_invalidated_at, created_at;

-- name: PromoteUserToAdminByEmail :one
UPDATE users
SET platform_role = 'admin'
WHERE LOWER(email) = LOWER($1)
  AND platform_role <> 'admin'
RETURNING id, email, username, platform_role, suspended_at, tokens_invalidated_at, created_at;

-- name: ListWorkspacesAdmin :many
SELECT
    id,
    name,
    slug,
    owner_id,
    created_at,
    updated_at
FROM workspaces
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetWorkspaceAdmin :one
SELECT
    id,
    name,
    slug,
    owner_id,
    created_at,
    updated_at
FROM workspaces
WHERE id = $1;

-- name: CountWorkspaceMembers :one
SELECT COUNT(*)::bigint AS count
FROM workspace_members
WHERE workspace_id = $1;

-- name: CountWorkspaceChannels :one
SELECT COUNT(*)::bigint AS count
FROM channels
WHERE workspace_id = $1;

-- name: ListAppsAdmin :many
SELECT
    a.id,
    a.name,
    a.slug,
    a.workspace_id,
    a.sandbox_workspace_id,
    a.owner_id,
    a.created_at,
    u.email AS owner_email
FROM apps a
INNER JOIN users u ON u.id = a.owner_id
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAppAdmin :one
SELECT
    a.id,
    a.name,
    a.slug,
    a.workspace_id,
    a.sandbox_workspace_id,
    a.owner_id,
    a.created_at,
    u.email AS owner_email
FROM apps a
INNER JOIN users u ON u.id = a.owner_id
WHERE a.id = $1;

-- name: ListAuditLogsAdmin :many
SELECT
    al.id,
    al.user_id,
    u.username AS username,
    al.app_id,
    al.action,
    al.method,
    al.path,
    al.ip_address,
    al.status_code,
    al.created_at
FROM audit_logs al
LEFT JOIN users u ON u.id = al.user_id
ORDER BY al.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogsAdminFiltered :many
SELECT
    al.id,
    al.user_id,
    u.username AS username,
    al.app_id,
    al.action,
    al.method,
    al.path,
    al.ip_address,
    al.status_code,
    al.created_at
FROM audit_logs al
LEFT JOIN users u ON u.id = al.user_id
WHERE (sqlc.narg('user_id')::bigint IS NULL OR al.user_id = sqlc.narg('user_id'))
  AND (COALESCE(sqlc.narg('action')::text, '') = '' OR al.action = sqlc.narg('action'))
ORDER BY al.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
