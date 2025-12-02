-- name: GetItem :one
select *
from item
where id = $1
;

-- name: GetItemByName :one
select *
from item
where name = $1
;

-- name: GetItemByFuzzy :one
select *
from item
order by levenshtein(name, $1) asc
limit 1
;


-- name: ListItem :many
select *
from item
;

-- name: CreateItem :one
INSERT INTO item (name, description, rarity, cost) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteItem :exec
delete from item
where id = $1
;

-- name: GetRandomItemByRarity :one
select *
from item
where rarity = $1
order by random()
limit 1
;

-- name: GetRandomItemByMinimumRarity :one
select *
from item
where rarity >= $1 and rarity != 'UNIQUE'
order by random()
limit 1
;

-- name: SearchItemByKeyword :many
SELECT * FROM item
WHERE LOWER(name) LIKE LOWER($1) OR LOWER(description) LIKE LOWER($1)
ORDER BY rarity DESC, name ASC
;

-- name: SearchItemByDescription :many
SELECT * FROM item
WHERE LOWER(description) LIKE LOWER($1)
ORDER BY rarity DESC, name ASC
;

