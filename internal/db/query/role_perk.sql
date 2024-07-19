-- name: CreateRolePerkJoin :exec
INSERT INTO role_perk (role_id, perk_id) VALUES ($1, $2);

-- name: ListRolePerkForRole :many
select perk_info.* from role_perk
inner join perk_info on role_perk.perk_id = perk_info.id
where role_perk.role_id = $1;
;


-- name: ListAssociatedRolesForPerk :many
select role.* from role_perk 
inner join role on role.id = role_perk.role_id
where role_perk.perk_id = $1
;
