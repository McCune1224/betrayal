-- name: CreatePlayerAbilityJoin :one
INSERT INTO player_ability (player_id, ability_id, quantity) VALUES ($1, $2, $3) RETURNING *;

