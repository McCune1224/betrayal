-- name: UpsertActionChannel :exec
insert into action_channel (channel_id) values ($1)
returning *;

-- name: GetActionChannel :one
select *
from action_channel
limit 1
;

-- name: WipeActionChannel :exec
delete from action_channel
;

