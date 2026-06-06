-- name: CreateWorkspace :one
INSERT INTO workspaces (
    name,
    slug,
    owner_id
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: AddWorkspaceMember :exec
INSERT INTO workspace_members (
    workspace_id,
    user_id,
    role
)
VALUES (
    $1,
    $2,
    $3
);

-- name: GetWorkspacesByUser :many
SELECT
    w.id,
    w.name,
    w.slug,
    w.owner_id,
    w.created_at
FROM workspaces w
INNER JOIN workspace_members wm
    ON w.id = wm.workspace_id
WHERE wm.user_id = $1
ORDER BY w.created_at DESC;

-- name: GetWorkspaceByID :one
SELECT
    id,
    name,
    slug,
    owner_id,
    created_at,
    updated_at
FROM workspaces
WHERE id = $1;

-- name: IsWorkspaceMember :one
SELECT EXISTS(
    SELECT 1
    FROM workspace_members
    WHERE workspace_id = $1 AND user_id = $2
)::boolean AS is_member;

-- name: GetWorkspaceBySlug :one
SELECT
    id,
    name,
    slug,
    owner_id,
    created_at,
    updated_at
FROM workspaces
WHERE slug = $1;

-- name: GetWorkspaceMember :one
SELECT
    id,
    workspace_id,
    user_id,
    role,
    created_at,
    updated_at
FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: GetWorkspaceMemberByUsername :one
SELECT
    wm.user_id,
    u.username,
    u.email
FROM workspace_members wm
INNER JOIN users u
    ON u.id = wm.user_id
WHERE wm.workspace_id = $1 AND LOWER(u.username) = LOWER($2);

-- name: ListWorkspaceMembers :many
SELECT
    wm.id,
    wm.workspace_id,
    wm.user_id,
    wm.role,
    wm.created_at,
    u.username,
    u.email
FROM workspace_members wm
INNER JOIN users u
    ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.created_at ASC;

-- name: UpdateWorkspaceMemberRole :exec
UPDATE workspace_members
SET role = $3, updated_at = NOW()
WHERE workspace_id = $1 AND user_id = $2;

-- name: RemoveWorkspaceMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;
