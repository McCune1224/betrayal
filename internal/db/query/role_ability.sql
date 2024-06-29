-- name: CreateRoleAbilityJoin :exec
INSERT INTO role_ability (role_id, ability_id) VALUES ($1, $2);
