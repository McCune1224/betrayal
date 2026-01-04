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

-- name: GetCommandActivitySummary :one
WITH last_hour AS (
    SELECT COUNT(*) AS total
    FROM command_audit
    WHERE timestamp >= NOW() - INTERVAL '1 hour'
),
last_day AS (
    SELECT
        COUNT(*) AS total,
        COUNT(*) FILTER (WHERE status = 'success') AS successes,
        COUNT(*) FILTER (WHERE status != 'success') AS failures,
        COUNT(*) FILTER (WHERE is_admin) AS admin_commands,
        COALESCE(AVG(execution_time_ms)::float8, 0::float8) AS avg_execution_time_ms
    FROM command_audit
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
)
SELECT
    COALESCE(last_hour.total, 0) AS commands_last_hour,
    COALESCE(last_day.total, 0) AS commands_last_24h,
    COALESCE(last_day.successes, 0) AS success_count_last_24h,
    COALESCE(last_day.failures, 0) AS failure_count_last_24h,
    COALESCE(last_day.admin_commands, 0) AS admin_commands_last_24h,
    COALESCE(last_day.avg_execution_time_ms, 0) AS avg_execution_time_ms_last_24h
FROM last_hour,
     last_day;

-- name: ListTopCommandsLastHour :many
SELECT
    command_name,
    COUNT(*) AS usage_count,
    COUNT(*) FILTER (WHERE status != 'success') AS failure_count,
    MAX(timestamp) AS last_used_at
FROM command_audit
WHERE timestamp >= NOW() - INTERVAL '1 hour'
GROUP BY command_name
ORDER BY usage_count DESC, command_name
LIMIT $1;

-- name: ListRecentCommandErrors :many
SELECT
    correlation_id,
    command_name,
    user_id,
    username,
    error_message,
    status,
    timestamp,
    command_arguments
FROM command_audit
WHERE status != 'success'
ORDER BY timestamp DESC
LIMIT $1;
