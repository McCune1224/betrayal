-- name: CreatePlayerStatusJoin :one
INSERT INTO player_status (player_id, status_id) VALUES ($1, $2) RETURNING *;

-- name: ListPlayerStatus :many
select status.*
from player_status
inner join status on player_status.status_id = status.id
where player_status.player_id = $1
;


-- name: ListPlayerStatusInventory :many
select status.*, player_status.quantity
from player_status
inner join status on player_status.status_id = status.id
where player_status.player_id = $1
;

