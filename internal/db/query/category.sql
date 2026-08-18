-- name: GetCategoryByName :one
select *
from category
where name = $1
;

-- name: GetCategory :one
select *
from category
where id = $1
;

-- name: GetCategoryByFuzzy :one
select *
from category
order by levenshtein(name, $1) asc
limit 1
;

-- name: ListCategory :many
select *
from category
;

-- name: CreateCategory :one
INSERT INTO category (name) VALUES ($1) RETURNING *;

-- name: UpdateCategory :one
UPDATE category
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteCategory :exec
delete from category
where id = $1
;