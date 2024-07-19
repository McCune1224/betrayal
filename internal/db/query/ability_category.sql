-- name: CreateAbilityCategoryJoin :exec
INSERT INTO ability_category (ability_id, category_id) VALUES ($1, $2);

-- name: ListAbilityCategoryNames :many
select category.name
from ability_category
inner join category on ability_category.category_id = category.id
where ability_category.ability_id = $1
;

