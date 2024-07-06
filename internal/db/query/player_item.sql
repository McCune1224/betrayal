-- name: UpsertPlayerItemJoin :exec
INSERT INTO player_item (player_id, item_id, quantity) VALUES ($1, $2, $3)
ON CONFLICT (player_id, item_id) 
DO UPDATE SET quantity = player_item.quantity + EXCLUDED.quantity
;

-- name: GetPlayerItem :one
select *
from item
inner join player_item on item.id = player_item.item_id
inner join player on player.id = player_item.player_id
where player.id = $1 and player_item.item_id = $2
;


-- name: ListPlayerItem :many
select item.*
from player_item
inner join item on player_item.item_id = item.id
where player_item.player_id = $1
;

-- name: ListPlayerItemInventory :many
select item.*, player_item.quantity
from player_item
inner join item on player_item.item_id = item.id
where player_item.player_id = $1
;


-- name: UpdatePlayerItemQuantity :one
UPDATE player_item SET quantity = $3 WHERE player_id = $1 AND item_id = $2 RETURNING *;

-- name: DeletePlayerItem :exec
delete from player_item
where player_id = $1 and item_id = $2
;

