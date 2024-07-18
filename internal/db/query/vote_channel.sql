-- name: UpsertVoteChannel :exec
insert into vote_channel (channel_id) values ($1)
returning *;

-- name: GetVoteChannel :one
select *
from vote_channel
limit 1
;

-- name: WipeVoteChannel :exec
delete from vote_channel
;

