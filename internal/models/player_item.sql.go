// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: player_item.sql

package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deletePlayerItem = `-- name: DeletePlayerItem :exec
delete from player_item
where player_id = $1 and item_id = $2
`

type DeletePlayerItemParams struct {
	PlayerID int64 `json:"player_id"`
	ItemID   int32 `json:"item_id"`
}

func (q *Queries) DeletePlayerItem(ctx context.Context, arg DeletePlayerItemParams) error {
	_, err := q.db.Exec(ctx, deletePlayerItem, arg.PlayerID, arg.ItemID)
	return err
}

const getPlayerItem = `-- name: GetPlayerItem :one
select item.id, name, description, rarity, cost, player_id, item_id, quantity, player.id, role_id, alive, coins, coin_bonus, luck, item_limit, alignment
from item
inner join player_item on item.id = player_item.item_id
inner join player on player.id = player_item.player_id
where player.id = $1 and player_item.item_id = $2
`

type GetPlayerItemParams struct {
	ID     int64 `json:"id"`
	ItemID int32 `json:"item_id"`
}

type GetPlayerItemRow struct {
	ID          int32          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Rarity      Rarity         `json:"rarity"`
	Cost        int32          `json:"cost"`
	PlayerID    int64          `json:"player_id"`
	ItemID      int32          `json:"item_id"`
	Quantity    int32          `json:"quantity"`
	ID_2        int64          `json:"id_2"`
	RoleID      pgtype.Int4    `json:"role_id"`
	Alive       bool           `json:"alive"`
	Coins       int32          `json:"coins"`
	CoinBonus   pgtype.Numeric `json:"coin_bonus"`
	Luck        int32          `json:"luck"`
	ItemLimit   int32          `json:"item_limit"`
	Alignment   Alignment      `json:"alignment"`
}

func (q *Queries) GetPlayerItem(ctx context.Context, arg GetPlayerItemParams) (GetPlayerItemRow, error) {
	row := q.db.QueryRow(ctx, getPlayerItem, arg.ID, arg.ItemID)
	var i GetPlayerItemRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Rarity,
		&i.Cost,
		&i.PlayerID,
		&i.ItemID,
		&i.Quantity,
		&i.ID_2,
		&i.RoleID,
		&i.Alive,
		&i.Coins,
		&i.CoinBonus,
		&i.Luck,
		&i.ItemLimit,
		&i.Alignment,
	)
	return i, err
}

const getPlayerItemCount = `-- name: GetPlayerItemCount :one
select coalesce(sum(quantity), 0) as item_count
from player_item
where player_item.player_id = $1
`

func (q *Queries) GetPlayerItemCount(ctx context.Context, playerID int64) (interface{}, error) {
	row := q.db.QueryRow(ctx, getPlayerItemCount, playerID)
	var item_count interface{}
	err := row.Scan(&item_count)
	return item_count, err
}

const listPlayerItem = `-- name: ListPlayerItem :many
select item.id, item.name, item.description, item.rarity, item.cost
from player_item
inner join item on player_item.item_id = item.id
where player_item.player_id = $1
`

func (q *Queries) ListPlayerItem(ctx context.Context, playerID int64) ([]Item, error) {
	rows, err := q.db.Query(ctx, listPlayerItem, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Rarity,
			&i.Cost,
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

const listPlayerItemInventory = `-- name: ListPlayerItemInventory :many
select item.id, item.name, item.description, item.rarity, item.cost, player_item.quantity
from player_item
inner join item on player_item.item_id = item.id
where player_item.player_id = $1
`

type ListPlayerItemInventoryRow struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rarity      Rarity `json:"rarity"`
	Cost        int32  `json:"cost"`
	Quantity    int32  `json:"quantity"`
}

func (q *Queries) ListPlayerItemInventory(ctx context.Context, playerID int64) ([]ListPlayerItemInventoryRow, error) {
	rows, err := q.db.Query(ctx, listPlayerItemInventory, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPlayerItemInventoryRow
	for rows.Next() {
		var i ListPlayerItemInventoryRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Rarity,
			&i.Cost,
			&i.Quantity,
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

const updatePlayerItemQuantity = `-- name: UpdatePlayerItemQuantity :one
UPDATE player_item SET quantity = $3 WHERE player_id = $1 AND item_id = $2 RETURNING player_id, item_id, quantity
`

type UpdatePlayerItemQuantityParams struct {
	PlayerID int64 `json:"player_id"`
	ItemID   int32 `json:"item_id"`
	Quantity int32 `json:"quantity"`
}

func (q *Queries) UpdatePlayerItemQuantity(ctx context.Context, arg UpdatePlayerItemQuantityParams) (PlayerItem, error) {
	row := q.db.QueryRow(ctx, updatePlayerItemQuantity, arg.PlayerID, arg.ItemID, arg.Quantity)
	var i PlayerItem
	err := row.Scan(&i.PlayerID, &i.ItemID, &i.Quantity)
	return i, err
}

const upsertPlayerItemJoin = `-- name: UpsertPlayerItemJoin :exec
INSERT INTO player_item (player_id, item_id, quantity) VALUES ($1, $2, $3)
ON CONFLICT (player_id, item_id) 
DO UPDATE SET quantity = player_item.quantity + EXCLUDED.quantity
`

type UpsertPlayerItemJoinParams struct {
	PlayerID int64 `json:"player_id"`
	ItemID   int32 `json:"item_id"`
	Quantity int32 `json:"quantity"`
}

func (q *Queries) UpsertPlayerItemJoin(ctx context.Context, arg UpsertPlayerItemJoinParams) error {
	_, err := q.db.Exec(ctx, upsertPlayerItemJoin, arg.PlayerID, arg.ItemID, arg.Quantity)
	return err
}
