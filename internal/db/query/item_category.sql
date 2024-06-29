-- name: CreateItemCategoryJoin :exec
INSERT INTO item_category (item_id, category_id) VALUES ($1, $2);
