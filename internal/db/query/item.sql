-- name: GetItem :one
SELECT * from item WHERE id = $1;

-- name: GetItemByName :one
SELECT * from item WHERE name = $1;

-- name: GetItemByFuzzy :one
SELECT * from item WHERE name ILIKE $1;

-- name: ListItem :many
SELECT * from item;

-- name: CreateItem :one
INSERT INTO item (name, description, rarity, cost) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteItem :exec
DELETE FROM item WHERE id = $1;
