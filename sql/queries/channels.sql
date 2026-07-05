-- name: CreateChannel :one
INSERT INTO channels (
    name,
    owner_id,
    workspace_id,
    is_read_only
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: AddChannelMember :exec
INSERT INTO channel_members (
    channel_id,
    user_id
)
VALUES (
    $1,
    $2
);

-- name: GetChannelsByWorkspaceAndUser :many
SELECT
    c.id,
    c.name,
    c.owner_id,
    c.workspace_id,
    c.is_read_only,
    c.created_at
FROM channels c
INNER JOIN channel_members cm
    ON c.id = cm.channel_id
WHERE cm.user_id = $1
    AND c.workspace_id = $2
ORDER BY c.created_at DESC;

-- name: IsChannelMember :one
SELECT EXISTS(
    SELECT 1
    FROM channel_members
    WHERE channel_id = $1 AND user_id = $2
)::boolean AS is_member;

-- name: GetChannelByID :one
SELECT
    id,
    name,
    owner_id,
    workspace_id,
    is_read_only,
    created_at,
    updated_at
FROM channels
WHERE id = $1;

-- name: GetChannelInWorkspace :one
SELECT
    id,
    name,
    owner_id,
    workspace_id,
    is_read_only,
    created_at,
    updated_at
FROM channels
WHERE id = $1
    AND workspace_id = $2;

-- name: GetChannelByNameInWorkspace :one
SELECT
    id,
    name,
    owner_id,
    workspace_id,
    is_read_only,
    created_at,
    updated_at
FROM channels
WHERE workspace_id = $1
    AND name = $2;

-- name: GetChannelMember :one
SELECT
    id,
    channel_id,
    user_id,
    created_at,
    updated_at
FROM channel_members
WHERE channel_id = $1 AND user_id = $2;

-- name: RemoveChannelMember :exec
DELETE FROM channel_members
WHERE channel_id = $1 AND user_id = $2;
