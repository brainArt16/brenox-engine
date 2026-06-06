-- name: CreateChannel :one
INSERT INTO channels (
    name,
    owner_id
)
VALUES (
    $1,
    $2
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

-- name: GetChannelsByUser :many
SELECT
    c.id,
    c.name,
    c.owner_id,
    c.created_at
FROM channels c
INNER JOIN channel_members cm
    ON c.id = cm.channel_id
WHERE cm.user_id = $1
ORDER BY c.created_at DESC;

-- name: IsChannelMember :one
SELECT EXISTS(
    SELECT 1
    FROM channel_members
    WHERE channel_id = $1 AND user_id = $2
)::boolean AS is_member;