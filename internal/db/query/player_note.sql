-- name: CreatePlayerNote :one
insert into player_note 
  (player_id, position, info) 
values ($1, $2, $3) 
returning *;

-- name: DeletePlayerNote :exec
delete from player_note
where player_id = $1 and note_id = $2
;

-- name: DeletePlayerNoteByPosition :exec
delete from player_note
where player_id = $1 and position = $2
;

-- name: GetPlayerNote :one
select *
from player_note
where player_id = $1 and note_id = $2
;

-- name: GetPlayerNoteByPosition :one
select *
from player_note
where player_id = $1 and position = $2
;

-- name: ListPlayerNote :many
select *
from player_note
where player_id = $1
;

-- name: GetPlayerNoteCount :one
select count(*) as note_count
from player_note
where player_id = $1
;

-- name: UpdatePlayerNoteByPosition :one
update player_note 
set info = $3, updated_at = now()
where player_id = $1 and position = $2
returning *
;

