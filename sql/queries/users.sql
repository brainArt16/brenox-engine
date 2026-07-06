-- name: CreateUser :one
INSERT INTO users (
    email,
    username,
    password_hash
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE LOWER(username) = LOWER($1);

-- name: UpdateUserProfile :one
UPDATE users
SET username = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: GetUserAuthState :one
SELECT
    id,
    email,
    platform_role,
    suspended_at,
    tokens_invalidated_at
FROM users
WHERE id = $1;
