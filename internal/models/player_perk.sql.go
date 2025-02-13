// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: player_perk.sql

package models

import (
	"context"
)

const createPlayerPerkJoin = `-- name: CreatePlayerPerkJoin :one
INSERT INTO player_perk (player_id, perk_id) VALUES ($1, $2) RETURNING player_id, perk_id
`

type CreatePlayerPerkJoinParams struct {
	PlayerID int64 `json:"player_id"`
	PerkID   int32 `json:"perk_id"`
}

func (q *Queries) CreatePlayerPerkJoin(ctx context.Context, arg CreatePlayerPerkJoinParams) (PlayerPerk, error) {
	row := q.db.QueryRow(ctx, createPlayerPerkJoin, arg.PlayerID, arg.PerkID)
	var i PlayerPerk
	err := row.Scan(&i.PlayerID, &i.PerkID)
	return i, err
}

const deletePlayerPerk = `-- name: DeletePlayerPerk :exec
delete from player_perk
where player_id = $1 and perk_id = $2
`

type DeletePlayerPerkParams struct {
	PlayerID int64 `json:"player_id"`
	PerkID   int32 `json:"perk_id"`
}

func (q *Queries) DeletePlayerPerk(ctx context.Context, arg DeletePlayerPerkParams) error {
	_, err := q.db.Exec(ctx, deletePlayerPerk, arg.PlayerID, arg.PerkID)
	return err
}

const listPlayerPerk = `-- name: ListPlayerPerk :many
select perk_info.id, perk_info.name, perk_info.description
from player_perk
inner join perk_info on player_perk.perk_id = perk_info.id
where player_perk.player_id = $1
`

func (q *Queries) ListPlayerPerk(ctx context.Context, playerID int64) ([]PerkInfo, error) {
	rows, err := q.db.Query(ctx, listPlayerPerk, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PerkInfo
	for rows.Next() {
		var i PerkInfo
		if err := rows.Scan(&i.ID, &i.Name, &i.Description); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
