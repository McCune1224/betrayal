-- name: GetPlayer :one
select *
from player
where id = $1
;

-- name: ListPlayer :many
select *
from player
;

-- name: CreatePlayer :one
INSERT INTO player (id, role_id, alive, coins, luck, alignment) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdatePlayer :one
UPDATE player SET role_id = $2, alive = $3, coins = $4, luck = $5, item_limit = $6, alignment = $7 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerAlive :one
UPDATE player SET alive = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerCoins :one
UPDATE player SET coins = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerAlignment :one
UPDATE player SET alignment = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerItemLimit :one
UPDATE player SET item_limit = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePlayerLuck :one
UPDATE player SET luck = $2 WHERE id = $1 RETURNING *;

-- name: DeletePlayer :exec
delete from player
where id = $1
;


-- name: GetPlayerInventory :one
select
    player.*,
    array_agg(distinct ability_info.*) as ability_details,
    array_agg(distinct item.*) as item_details,
    array_agg(distinct "status".*) as immunity_details
from player
inner join player_ability on player.id = player_ability.player_id
inner join ability_info on player_ability.ability_id = ability_info.id
left join player_item on player.id = player_item.player_id
left join item on player_item.item_id = item.id
left join player_immunity on player.id = player_immunity.player_id
left join status on player_immunity.status_id = status.id
where player.id = $1
group by player.id
;


-- name: PlayerFoo :one
select
    player.id,
    player.role_id,
    player.alive,
    player.coins,
    player.luck,
    player.item_limit,
    player.alignment,
    json_agg(
        json_build_object(
            'ability_id',
            ability_info.id,
            'name',
            ability_info.name,
            'description',
            ability_info.description
        )
    ) as ability_details,
    json_agg(
        json_build_object(
            'item_id', item.id, 'name', item.name, 'description', item.description
        )
    ) as item_details,
    json_agg(
        json_build_object(
            'immunity_id',
            status.id,
            'name',
            status.name,
            'description',
            status.description
        )
    ) as immunity_details
from player
left join player_ability on player.id = player_ability.player_id
left join ability_info on player_ability.ability_id = ability_info.id
left join player_item on player.id = player_item.player_id
left join item on player_item.item_id = item.id
left join player_immunity on player.id = player_immunity.player_id
left join status on player_immunity.status_id = status.id
where player.id = $1
group by player.id
;

