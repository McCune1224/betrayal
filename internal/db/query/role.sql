-- name: GetRole :one
SELECT * FROM role WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM role WHERE name = $1;

-- name: GetRoleByFuzzy :one
SELECT * FROM role WHERE name ILIKE $1;

-- name: Listrole :many
SELECT * FROM role;


-- name: CreateRole :one
INSERT INTO role (name, description, alignment) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteRole :exec
DELETE FROM role WHERE id = $1;

-- name: NukeRoles :exec
TRUNCATE TABLE role, role_ability, role_perk, role_ability, role_perk, ability_category RESTART IDENTITY CASCADE;
