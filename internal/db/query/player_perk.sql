-- name: CreatePlayerPerkJoin :one
INSERT INTO player_perk (player_id, perk_id) VALUES ($1, $2) RETURNING *;

-- name: ListPlayerPerk :many
select perk_info.*
from player_perk
inner join perk_info on player_perk.perk_id = perk_info.id
where player_perk.player_id = $1
;

