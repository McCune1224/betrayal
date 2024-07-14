-- name: GetAbilityInfo :one
select *
from ability_info
where id = $1
;

-- name: GetAbilityInfoByName :one
select *
from ability_info
where name = $1
;

-- name: GetAbilityInfoByFuzzy :one
select *
from ability_info
order by levenshtein(name, $1) asc
limit 1
;

-- name: GetAnyAbilityByFuzzy :one
select *
from ability_info
where ability_info.any_ability = true
order by levenshtein(name, $1) asc
limit 1
;

-- name: ListAbilityInfo :many
select *
from ability_info
;

-- name: CreateAbilityInfo :one
INSERT INTO ability_info (name, description, default_charges, any_ability, rarity) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteAbilityInfo :exec
delete from ability_info
where id = $1
;

