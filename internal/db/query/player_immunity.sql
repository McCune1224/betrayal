-- name: CreatePlayerImmunityJoin :one
INSERT INTO player_immunity (player_id, status_id) VALUES ($1, $2) RETURNING *;

-- name: ListPlayerImmunity :many
select status.*
from player_immunity
inner join status on status.id = player_immunity.status_id
where player_immunity.player_id = $1
;

