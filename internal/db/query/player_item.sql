-- name: UpsertPlayerItemJoin :exec
INSERT INTO player_item (player_id, item_id, quantity) VALUES ($1, $2, $3)
ON CONFLICT (player_id, item_id) 
DO UPDATE SET quantity = player_item.quantity + EXCLUDED.quantity
;

-- name: GetPlayerItem :one
SELECT * from item 
INNER JOIN player_item ON item.id = player_item.item_id
INNER JOIN player ON player.id = player_item.player_id
WHERE player.id = $1 AND player_item.item_id = $2;


-- name: ListPlayerItem :many
SELECT * from item 
INNER JOIN player_item ON item.id = player_item.item_id
INNER JOIN player ON player.id = player_item.player_id
WHERE player.id = $1;


-- name: UpdatePlayerItemQuantity :one
UPDATE player_item SET quantity = $3 WHERE player_id = $1 AND item_id = $2 RETURNING *;

-- name: DeletePlayerItem :exec
DELETE FROM player_item WHERE player_id = $1 AND item_id = $2;
