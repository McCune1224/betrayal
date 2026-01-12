-- name: GetRole :one
select *
from role
where id = $1
;

-- name: GetRoleByName :one
select *
from role
where name = $1
;

-- name: ListRolesByName :many
select *
from role
where name = any($1::text[])
;

-- name: GetRoleByFuzzy :one
select *
from role
order by levenshtein(name, $1) asc
limit 1
;

-- name: Listrole :many
select *
from role
;


-- name: CreateRole :one
INSERT INTO role (name, description, alignment) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteRole :exec
delete from role
where id = $1
;

-- name: NukeRoles :exec
TRUNCATE TABLE role, role_ability, role_perk, role_ability, role_perk, ability_category RESTART IDENTITY CASCADE;

-- name: UpdateRole :one
UPDATE role
SET name = $2, description = $3, alignment = $4
WHERE id = $1
RETURNING *;

-- name: SearchRoleByName :many
SELECT *
FROM role
ORDER BY levenshtein(LOWER(name), LOWER($1)) ASC
LIMIT 20;
