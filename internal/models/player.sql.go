// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: player.sql

package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPlayer = `-- name: CreatePlayer :one
INSERT INTO player (id, role_id, alive, coins, luck, alignment) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type CreatePlayerParams struct {
	ID        int64       `json:"id"`
	RoleID    pgtype.Int4 `json:"role_id"`
	Alive     bool        `json:"alive"`
	Coins     int32       `json:"coins"`
	Luck      int32       `json:"luck"`
	Alignment Alignment   `json:"alignment"`
}

func (q *Queries) CreatePlayer(ctx context.Context, arg CreatePlayerParams) (Player, error) {
	row := q.db.QueryRow(ctx, createPlayer,
		arg.ID,
		arg.RoleID,
		arg.Alive,
		arg.Coins,
		arg.Luck,
		arg.Alignment,
	)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const deletePlayer = `-- name: DeletePlayer :exec
delete from player
where id = $1
`

func (q *Queries) DeletePlayer(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deletePlayer, id)
	return err
}

const getPlayer = `-- name: GetPlayer :one
select id, role_id, alive, coins, luck, item_limit, alignment
from player
where id = $1
`

func (q *Queries) GetPlayer(ctx context.Context, id int64) (Player, error) {
	row := q.db.QueryRow(ctx, getPlayer, id)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const listPlayer = `-- name: ListPlayer :many
select id, role_id, alive, coins, luck, item_limit, alignment
from player
`

func (q *Queries) ListPlayer(ctx context.Context) ([]Player, error) {
	rows, err := q.db.Query(ctx, listPlayer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Player
	for rows.Next() {
		var i Player
		if err := rows.Scan(
			&i.ID,
			&i.RoleID,
			&i.Alive,
			&i.Coins,
			&i.Luck,
			&i.ItemLimit,
			&i.Alignment,
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

const updatePlayer = `-- name: UpdatePlayer :one
UPDATE player SET role_id = $2, alive = $3, coins = $4, luck = $5, item_limit = $6, alignment = $7 WHERE id = $1 RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type UpdatePlayerParams struct {
	ID        int64       `json:"id"`
	RoleID    pgtype.Int4 `json:"role_id"`
	Alive     bool        `json:"alive"`
	Coins     int32       `json:"coins"`
	Luck      int32       `json:"luck"`
	ItemLimit int32       `json:"item_limit"`
	Alignment Alignment   `json:"alignment"`
}

func (q *Queries) UpdatePlayer(ctx context.Context, arg UpdatePlayerParams) (Player, error) {
	row := q.db.QueryRow(ctx, updatePlayer,
		arg.ID,
		arg.RoleID,
		arg.Alive,
		arg.Coins,
		arg.Luck,
		arg.ItemLimit,
		arg.Alignment,
	)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const updatePlayerAlignment = `-- name: UpdatePlayerAlignment :one
UPDATE player SET alignment = $2 WHERE id = $1 RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type UpdatePlayerAlignmentParams struct {
	ID        int64     `json:"id"`
	Alignment Alignment `json:"alignment"`
}

func (q *Queries) UpdatePlayerAlignment(ctx context.Context, arg UpdatePlayerAlignmentParams) (Player, error) {
	row := q.db.QueryRow(ctx, updatePlayerAlignment, arg.ID, arg.Alignment)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const updatePlayerAlive = `-- name: UpdatePlayerAlive :one
UPDATE player SET alive = $2 WHERE id = $1 RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type UpdatePlayerAliveParams struct {
	ID    int64 `json:"id"`
	Alive bool  `json:"alive"`
}

func (q *Queries) UpdatePlayerAlive(ctx context.Context, arg UpdatePlayerAliveParams) (Player, error) {
	row := q.db.QueryRow(ctx, updatePlayerAlive, arg.ID, arg.Alive)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const updatePlayerCoins = `-- name: UpdatePlayerCoins :one
UPDATE player SET coins = $2 WHERE id = $1 RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type UpdatePlayerCoinsParams struct {
	ID    int64 `json:"id"`
	Coins int32 `json:"coins"`
}

func (q *Queries) UpdatePlayerCoins(ctx context.Context, arg UpdatePlayerCoinsParams) (Player, error) {
	row := q.db.QueryRow(ctx, updatePlayerCoins, arg.ID, arg.Coins)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const updatePlayerItemLimit = `-- name: UpdatePlayerItemLimit :one
UPDATE player SET item_limit = $2 WHERE id = $1 RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type UpdatePlayerItemLimitParams struct {
	ID        int64 `json:"id"`
	ItemLimit int32 `json:"item_limit"`
}

func (q *Queries) UpdatePlayerItemLimit(ctx context.Context, arg UpdatePlayerItemLimitParams) (Player, error) {
	row := q.db.QueryRow(ctx, updatePlayerItemLimit, arg.ID, arg.ItemLimit)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const updatePlayerLuck = `-- name: UpdatePlayerLuck :one
UPDATE player SET luck = $2 WHERE id = $1 RETURNING id, role_id, alive, coins, luck, item_limit, alignment
`

type UpdatePlayerLuckParams struct {
	ID   int64 `json:"id"`
	Luck int32 `json:"luck"`
}

func (q *Queries) UpdatePlayerLuck(ctx context.Context, arg UpdatePlayerLuckParams) (Player, error) {
	row := q.db.QueryRow(ctx, updatePlayerLuck, arg.ID, arg.Luck)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}