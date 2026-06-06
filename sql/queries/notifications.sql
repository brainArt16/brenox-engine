-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    type,
    title,
    body,
    data
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetNotificationByID :one
SELECT *
FROM notifications
WHERE id = $1;

-- name: ListNotificationsByUser :many
SELECT *
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :one
UPDATE notifications
SET read_at = NOW()
WHERE id = $1 AND user_id = $2 AND read_at IS NULL
RETURNING *;

-- name: MarkAllNotificationsRead :execrows
UPDATE notifications
SET read_at = NOW()
WHERE user_id = $1 AND read_at IS NULL;
