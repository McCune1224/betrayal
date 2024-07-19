-- name: CreatePlayerAbilityJoin :one
INSERT INTO player_ability (player_id, ability_id, quantity) VALUES ($1, $2, $3) RETURNING *;


-- name: ListPlayerAbility :many
select ability_info.*
from player_ability
inner join ability_info on player_ability.ability_id = ability_info.id
where player_ability.player_id = $1
;


-- name: ListPlayerAbilityJoin :many
select player_ability.*
from player_ability
where player_ability.player_id = $1
;


-- name: UpdatePlayerAbilityQuantity :one
update player_ability set quantity = $1 where player_ability.player_id = $2 and player_ability.ability_id = $3
returning *;
;


-- name: DeletePlayerAbility :exec
delete from player_ability
where player_ability.player_id = $1 and player_ability.ability_id = $2
;

-- name: ListPlayerAbilityInventory :many
select ability_info.*, player_ability.quantity
from player_ability
inner join ability_info on player_ability.ability_id = ability_info.id
where player_ability.player_id = $1
;

