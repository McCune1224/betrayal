-- name: CreatePlayerConfessional :one
INSERT INTO player_confessional (player_id, channel_id) VALUES ($1, $2) RETURNING *;

