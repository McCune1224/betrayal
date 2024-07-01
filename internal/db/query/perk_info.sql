-- name: GetPerkInfo :one
select *
from perk_info
where id = $1
;

-- name: GetPerkInfoByName :one
select *
from perk_info
where name = $1
;

-- name: GetPerkInfoByFuzzy :one
select *
from perk_info
order by levenshtein(name, $1) asc
limit 1
;

-- name: ListPerkInfo :many
select *
from perk_info
;

-- name: CreatePerkInfo :one
INSERT INTO perk_info (name, description) VALUES ($1, $2) RETURNING *;

-- name: DeletePerkInfo :exec
delete from perk_info
where id = $1
;

