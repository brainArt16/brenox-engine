-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    user_id,
    app_id,
    action,
    method,
    path,
    ip_address,
    status_code
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
);
