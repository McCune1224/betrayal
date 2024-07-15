-- name: CreateRoleAbilityJoin :exec
insert into role_ability (role_id, ability_id) values ($1, $2);


-- name: ListRoleAbilityForRole :many
select ability_info.*
from role_ability
inner join ability_info on role_ability.ability_id = ability_info.id
where role_ability.role_id = $1
;
;


-- name: ListAssociatedRolesForAbility :many
select role.*
from role_ability
inner join role on role.id = role_ability.role_id
where role_ability.ability_id = $1
;


-- name: RollAbility :one
select *
from ability_info
inner join role_ability on role_ability.ability_id = $1
where ability_info.any_ability = true

