-- name: CreatePlayerLifeboard :one
insert into player_lifeboard 
  (channel_id, message_id) VALUES ($1, $2) RETURNING *;

-- name: GetPlayerLifeboard :one
select *
from player_lifeboard
limit 1
;

-- name: DeletePlayerLifeboard :exec
delete from player_lifeboard
;

