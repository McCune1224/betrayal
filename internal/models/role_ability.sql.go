// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: role_ability.sql

package models

import (
	"context"
)

const createRoleAbilityJoin = `-- name: CreateRoleAbilityJoin :exec
insert into role_ability (role_id, ability_id) values ($1, $2)
`

type CreateRoleAbilityJoinParams struct {
	RoleID    int32 `json:"role_id"`
	AbilityID int32 `json:"ability_id"`
}

func (q *Queries) CreateRoleAbilityJoin(ctx context.Context, arg CreateRoleAbilityJoinParams) error {
	_, err := q.db.Exec(ctx, createRoleAbilityJoin, arg.RoleID, arg.AbilityID)
	return err
}

const listRoleAbilityForRole = `-- name: ListRoleAbilityForRole :many
select ability_info.id, ability_info.name, ability_info.description, ability_info.default_charges, ability_info.any_ability, ability_info.rarity from role_ability
inner join ability_info on role_ability.ability_id = ability_info.id
where role_ability.role_id = $1
`

func (q *Queries) ListRoleAbilityForRole(ctx context.Context, roleID int32) ([]AbilityInfo, error) {
	rows, err := q.db.Query(ctx, listRoleAbilityForRole, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AbilityInfo
	for rows.Next() {
		var i AbilityInfo
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.DefaultCharges,
			&i.AnyAbility,
			&i.Rarity,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
