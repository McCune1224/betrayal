-- name: CreateItemCategoryJoin :exec
INSERT INTO item_category (item_id, category_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListItemCategoryNames :many
select category.name
from item_category
inner join category on item_category.category_id = category.id
where item_category.item_id = $1
;

-- name: DeleteItemCategoryJoin :exec
DELETE FROM item_category
WHERE item_id = $1 AND category_id = $2;
