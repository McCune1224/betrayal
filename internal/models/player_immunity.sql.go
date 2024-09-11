// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: player_immunity.sql

package models

import (
	"context"
)

const createOneTimePlayerImmunityJoin = `-- name: CreateOneTimePlayerImmunityJoin :one
INSERT INTO player_immunity (player_id, status_id, one_time) VALUES ($1, $2, $3) RETURNING player_id, status_id, one_time
`

type CreateOneTimePlayerImmunityJoinParams struct {
	PlayerID int64 `json:"player_id"`
	StatusID int32 `json:"status_id"`
	OneTime  bool  `json:"one_time"`
}

func (q *Queries) CreateOneTimePlayerImmunityJoin(ctx context.Context, arg CreateOneTimePlayerImmunityJoinParams) (PlayerImmunity, error) {
	row := q.db.QueryRow(ctx, createOneTimePlayerImmunityJoin, arg.PlayerID, arg.StatusID, arg.OneTime)
	var i PlayerImmunity
	err := row.Scan(&i.PlayerID, &i.StatusID, &i.OneTime)
	return i, err
}

const createPlayerImmunityJoin = `-- name: CreatePlayerImmunityJoin :one
INSERT INTO player_immunity (player_id, status_id) VALUES ($1, $2) RETURNING player_id, status_id, one_time
`

type CreatePlayerImmunityJoinParams struct {
	PlayerID int64 `json:"player_id"`
	StatusID int32 `json:"status_id"`
}

func (q *Queries) CreatePlayerImmunityJoin(ctx context.Context, arg CreatePlayerImmunityJoinParams) (PlayerImmunity, error) {
	row := q.db.QueryRow(ctx, createPlayerImmunityJoin, arg.PlayerID, arg.StatusID)
	var i PlayerImmunity
	err := row.Scan(&i.PlayerID, &i.StatusID, &i.OneTime)
	return i, err
}

const deletePlayerImmunity = `-- name: DeletePlayerImmunity :exec
delete from player_immunity
where player_id = $1 and status_id = $2
`

type DeletePlayerImmunityParams struct {
	PlayerID int64 `json:"player_id"`
	StatusID int32 `json:"status_id"`
}

func (q *Queries) DeletePlayerImmunity(ctx context.Context, arg DeletePlayerImmunityParams) error {
	_, err := q.db.Exec(ctx, deletePlayerImmunity, arg.PlayerID, arg.StatusID)
	return err
}

const listPlayerImmunity = `-- name: ListPlayerImmunity :many
select status.id, status.name, status.description, status.hour_duration
from player_immunity
inner join status on status.id = player_immunity.status_id
where player_immunity.player_id = $1
`

func (q *Queries) ListPlayerImmunity(ctx context.Context, playerID int64) ([]Status, error) {
	rows, err := q.db.Query(ctx, listPlayerImmunity, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Status
	for rows.Next() {
		var i Status
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.HourDuration,
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
