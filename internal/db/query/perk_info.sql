-- name: GetPerkInfo :one
SELECT * from perk_info WHERE id = $1;

-- name: GetPerkInfoByName :one
SELECT * from perk_info WHERE name = $1;

-- name: GetPerkInfoByFuzzy :one
SELECT * from perk_info WHERE name ILIKE $1;

-- name: ListPerkInfo :many
SELECT * from perk_info;

-- name: CreatePerkInfo :one
INSERT INTO perk_info (name, description) VALUES ($1, $2) RETURNING *;

-- name: DeletePerkInfo :exec
DELETE FROM perk_info WHERE id = $1;
