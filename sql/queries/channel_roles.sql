-- name: GetChannelRole :one
SELECT
    id,
    channel_id,
    user_id,
    role,
    created_at
FROM channel_roles
WHERE channel_id = $1 AND user_id = $2;

-- name: UpsertChannelRole :exec
INSERT INTO channel_roles (
    channel_id,
    user_id,
    role
)
VALUES (
    $1,
    $2,
    $3
)
ON CONFLICT (channel_id, user_id)
DO UPDATE SET role = EXCLUDED.role;
