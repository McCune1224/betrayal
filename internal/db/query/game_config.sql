-- name: GetGameConfig :one
SELECT value
FROM game_config
WHERE key = $1;

-- name: UpsertGameConfig :one
INSERT INTO game_config (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = NOW()
RETURNING *;

-- name: ListGameConfig :many
SELECT * FROM game_config
ORDER BY key;

-- name: DeleteGameConfig :exec
DELETE FROM game_config
WHERE key = $1;
