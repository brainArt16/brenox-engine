-- name: CreateAttachment :one
INSERT INTO attachments (
    message_id,
    uploader_id,
    object_key,
    file_name,
    mime_type,
    size_bytes
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: ListAttachmentsByMessage :many
SELECT *
FROM attachments
WHERE message_id = $1
ORDER BY created_at ASC;

-- name: GetAttachmentByID :one
SELECT *
FROM attachments
WHERE id = $1;

-- name: GetMessageInChannel :one
SELECT
    m.id,
    m.channel_id,
    m.sender_id,
    m.content,
    m.created_at
FROM messages m
INNER JOIN channels c
    ON c.id = m.channel_id
WHERE m.id = $1
  AND m.channel_id = $2
  AND c.workspace_id = $3;
