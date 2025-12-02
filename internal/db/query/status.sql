-- name: GetStatusByName :one
select *
from status
where name = $1
;

-- name: GetStatusByFuzzy :one
select *
from status
order by levenshtein(name, $1) asc
limit 1
;

-- name: ListStatus :many
select *
from status
;

-- name: CreateStatus :one
INSERT INTO status (name, description) VALUES ($1, $2) RETURNING *;

-- name: DeleteStatus :exec
delete from status
where id = $1
;

-- name: SearchStatusByKeyword :many
SELECT * FROM status
WHERE LOWER(name) LIKE LOWER($1) OR LOWER(description) LIKE LOWER($1)
ORDER BY name ASC
;

-- name: SearchStatusByDescription :many
SELECT * FROM status
WHERE LOWER(description) LIKE LOWER($1)
ORDER BY name ASC
;

