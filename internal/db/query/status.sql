-- name: GetStatusByName :one
SELECT * from status WHERE name = $1;

-- name: GetStatusByFuzzy :one
SELECT * from status WHERE name ILIKE $1;

-- name: ListStatus :many
SELECT * from status;

-- name: CreateStatus :one
INSERT INTO status (name, description) VALUES ($1, $2) RETURNING *;

-- name: DeleteStatus :exec
DELETE FROM status WHERE id = $1;
