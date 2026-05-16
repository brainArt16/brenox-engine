-- name: CreateMessage :one
INSERT INTO messages (
    channel_id,
    sender_id,
    content
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetChannelMessages :many
SELECT
    m.id,
    m.channel_id,
    m.sender_id,
    m.content,
    m.created_at,
    u.username
FROM messages m
INNER JOIN users u
    ON m.sender_id = u.id
WHERE m.channel_id = $1
ORDER BY m.created_at ASC
LIMIT $2 OFFSET $3;