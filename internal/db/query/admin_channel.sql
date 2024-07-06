-- name: ListAdminChannel :many
select *
from admin_channel
;

-- name: CreateAdminChannel :one
INSERT INTO admin_channel 
(channel_id) 
VALUES ($1) 
returning *; 

-- name: DeleteAdminChannel :exec
delete from admin_channel
where channel_id = $1
;

