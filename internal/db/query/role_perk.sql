-- name: CreateRolePerkJoin :exec
INSERT INTO role_perk (role_id, perk_id) VALUES ($1, $2);
