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

-- name: ListAnyAbilities :many
select *
from ability_info
where ability_info.any_ability = true and ability_info.rarity != 'ROLE_SPECIFIC'
;

-- name: ListAnyAbilitiesIncludingRoleSpecific :many
select distinct ability_info.*
from role_ability
inner join ability_info on ability_info.id = role_ability.ability_id
where
    (ability_info.any_ability = true and ability_info.rarity != 'ROLE_SPECIFIC')
    or (role_ability.role_id = $1 and ability_info.any_ability = true)
;


-- name: GetRandomAnyAbilityIncludingRoleSpecific :one
select ability_info.*
from role_ability
inner join ability_info on ability_info.id = role_ability.ability_id
where
    (ability_info.any_ability = true and ability_info.rarity = $1)
    or (role_ability.role_id = $2 and ability_info.any_ability = true)
order by random()
limit 1
;


-- name: GetRandomAnyAbilityByRarity :one
select *
from ability_info
where ability_info.any_ability = true and ability_info.rarity == $1
;

-- name: GetRandomAnyAbilityByMinimumRarity :one
select *
from ability_info
where
    ability_info.any_ability = true
    and ability_info.rarity >= $1
    and ability_info.rarity != 'ROLE_SPECIFIC'
;

