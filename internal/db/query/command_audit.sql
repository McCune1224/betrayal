-- name: InsertCommandAudit :exec
INSERT INTO command_audit (
    correlation_id,
    command_name,
    user_id,
    username,
    user_roles,
    guild_id,
    channel_id,
    is_admin,
    command_arguments,
    status,
    error_message,
    execution_time_ms,
    environment
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
);

-- name: ListCommandAuditByUser :many
SELECT *
FROM command_audit
WHERE user_id = $1
ORDER BY timestamp DESC
LIMIT $2;

-- name: ListCommandAuditByCommand :many
SELECT *
FROM command_audit
WHERE command_name = $1
ORDER BY timestamp DESC
LIMIT $2;

-- name: ListAdminCommands :many
SELECT *
FROM command_audit
WHERE is_admin = true
ORDER BY timestamp DESC
LIMIT $1;

-- name: GetCommandAuditByCorrelationID :one
SELECT *
FROM command_audit
WHERE correlation_id = $1;

-- name: ListRecentCommands :many
SELECT *
FROM command_audit
WHERE timestamp > NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC
LIMIT 100;
