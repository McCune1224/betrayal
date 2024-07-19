-- name: CreatePlayerConfessional :one
INSERT INTO player_confessional (player_id, channel_id, pin_message_id) VALUES ($1, $2, $3) RETURNING *;

-- name: GetPlayerConfessional :one
select *
from player_confessional
where player_id = $1
;


-- name: GetPlayerConfessionalByChannelID :one
select *
from player_confessional
where channel_id = $1
;

-- name: ListPlayerConfessional :many
select *
from player_confessional
;

-- name: DeletePlayerConfessional :exec
delete from player_confessional
where player_id = $1
;

