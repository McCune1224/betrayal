-- name: CreatePlayerPerkJoin :one
INSERT INTO player_perk (player_id, perk_id) VALUES ($1, $2) RETURNING *;

