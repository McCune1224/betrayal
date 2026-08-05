-- name: GetCommandLogChannel :one
SELECT channel_id
FROM command_log_channel
WHERE id = 1;

-- name: SetCommandLogChannel :one
INSERT INTO command_log_channel (id, channel_id)
VALUES (1, $1)
ON CONFLICT (id) DO UPDATE SET
    channel_id = EXCLUDED.channel_id,
    updated_at = NOW()
RETURNING channel_id;

-- name: DeleteCommandLogChannel :exec
DELETE FROM command_log_channel
WHERE id = 1;
