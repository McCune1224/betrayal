-- name: GetPlayer :one
SELECT * from player WHERE id = $1;

-- name: ListPlayer :many
SELECT * from player;

-- name: CreatePlayer :one
INSERT INTO player (role_id, alive, coins, luck, alignment) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdatePlayer :one
UPDATE player SET role_id = $2, alive = $3, coins = $4, luck = $5, alignment = $6 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerAlive :one
UPDATE player SET alive = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerCoins :one
UPDATE player SET coins = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerAlignment :one
UPDATE player SET alignment = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerLuck :one
UPDATE player SET luck = $2 WHERE id = $1 RETURNING *;

-- name: DeletePlayer :exec
DELETE FROM player WHERE id = $1;
