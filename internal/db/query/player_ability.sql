-- name: CreatePlayerAbilityJoin :one
INSERT INTO player_ability (player_id, ability_id, quantity) VALUES ($1, $2, $3) RETURNING *;

-- name: ListPlayerAbility :many
select ability_info.*
from player_ability
inner join ability_info on player_ability.ability_id = ability_info.id
where player_ability.player_id = $1
;

