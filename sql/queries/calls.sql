-- name: CreateCall :one
INSERT INTO calls (
    channel_id,
    workspace_id,
    initiator_id,
    status
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetCallByID :one
SELECT *
FROM calls
WHERE id = $1;

-- name: GetActiveCallByChannel :one
SELECT *
FROM calls
WHERE channel_id = $1
  AND status IN ('ringing', 'active')
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateCallStatus :one
UPDATE calls
SET status = $2,
    ended_at = CASE WHEN $2 = 'ended' THEN NOW() ELSE ended_at END
WHERE id = $1
RETURNING *;

-- name: AddCallParticipant :one
INSERT INTO call_participants (
    call_id,
    user_id
)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetActiveCallParticipant :one
SELECT *
FROM call_participants
WHERE call_id = $1
  AND user_id = $2
  AND left_at IS NULL;

-- name: ListActiveCallParticipants :many
SELECT *
FROM call_participants
WHERE call_id = $1
  AND left_at IS NULL
ORDER BY joined_at ASC;

-- name: CountActiveCallParticipants :one
SELECT COUNT(*)::bigint AS count
FROM call_participants
WHERE call_id = $1
  AND left_at IS NULL;

-- name: MarkCallParticipantLeft :one
UPDATE call_participants
SET left_at = NOW()
WHERE call_id = $1
  AND user_id = $2
  AND left_at IS NULL
RETURNING *;

-- name: ListChannelMemberUserIDs :many
SELECT user_id
FROM channel_members
WHERE channel_id = $1;
