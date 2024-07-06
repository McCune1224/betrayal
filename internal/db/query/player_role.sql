-- name: CreatePlayerRoleJoin :one
INSERT INTO player_role (player_id, role_id) VALUES ($1, $2) RETURNING *;

-- name: GetPlayerRole :one
select role.*
from player_role
inner join role on player_role.role_id = role.id
where player_role.player_id = $1
;

