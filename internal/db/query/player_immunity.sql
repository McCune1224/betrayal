-- name: CreatePlayerImmunityJoin :one
INSERT INTO player_immunity (player_id, status_id) VALUES ($1, $2) RETURNING *;

-- name: CreateOneTimePlayerImmunityJoin :one
INSERT INTO player_immunity (player_id, status_id, one_time) VALUES ($1, $2, $3) RETURNING *;

-- name: ListPlayerImmunity :many
select status.*, player_immunity.one_time
from player_immunity
inner join status on status.id = player_immunity.status_id
where player_immunity.player_id = $1
;

-- name: DeletePlayerImmunity :exec
delete from player_immunity
where player_id = $1 and status_id = $2
;

