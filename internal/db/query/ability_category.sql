-- name: CreateAbilityCategoryJoin :exec
INSERT INTO ability_category (ability_id, category_id) VALUES ($1, $2);
