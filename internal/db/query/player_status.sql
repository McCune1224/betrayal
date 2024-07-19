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

-- name: UpsertPlayerStatusJoin :exec
INSERT INTO player_status (player_id, status_id, quantity) VALUES ($1, $2, $3)
ON CONFLICT (player_id, status_id)
DO UPDATE SET quantity = player_status.quantity + EXCLUDED.quantity
;


-- name: DeletePlayerStatus :exec
delete from player_status
where player_id = $1 and status_id = $2
;


-- name: UpdatePlayerStatusQuantity :one
UPDATE player_status SET quantity = $3 WHERE player_id = $1 AND status_id = $2 RETURNING *;

