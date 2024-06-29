-- name: GetAbilityInfo :one
SELECT * from ability_info WHERE id = $1;

-- name: GetAbilityInfoByName :one
SELECT * from ability_info WHERE name = $1;

-- name: GetAbilityInfoByFuzzy :one
SELECT * from ability_info WHERE name ILIKE $1;

-- name: ListAbilityInfo :many
SELECT * from ability_info;

-- name: CreateAbilityInfo :one
INSERT INTO ability_info (name, description, default_charges, any_ability, rarity) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteAbilityInfo :exec
DELETE FROM ability_info WHERE id = $1;
