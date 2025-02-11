-- name: GetCycle :one
select *
from game_cycle
limit 1;

-- name: UpdateCycle :one
update game_cycle
set is_elimination = $1,
day = $2
WHERE id = $3
returning *;
