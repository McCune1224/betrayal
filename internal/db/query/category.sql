-- name: GetCategoryByName :one
SELECT * from category WHERE name = $1;

-- name: GetCategoryByFuzzy :one
SELECT * from category WHERE name ILIKE $1;

-- name: ListCategory :many
SELECT * from category;

-- name: CreateCategory :one
INSERT INTO category (name) VALUES ($1) RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM category WHERE id = $1;
