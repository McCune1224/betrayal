// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: action_channel.sql

package models

import (
	"context"
)

const getActionChannel = `-- name: GetActionChannel :one
select channel_id
from action_channel
limit 1
`

func (q *Queries) GetActionChannel(ctx context.Context) (string, error) {
	row := q.db.QueryRow(ctx, getActionChannel)
	var channel_id string
	err := row.Scan(&channel_id)
	return channel_id, err
}

const upsertActionChannel = `-- name: UpsertActionChannel :exec
insert into action_channel (channel_id) values ($1)
returning channel_id
`

func (q *Queries) UpsertActionChannel(ctx context.Context, channelID string) error {
	_, err := q.db.Exec(ctx, upsertActionChannel, channelID)
	return err
}

const wipeActionChannel = `-- name: WipeActionChannel :exec
delete from action_channel
`

func (q *Queries) WipeActionChannel(ctx context.Context) error {
	_, err := q.db.Exec(ctx, wipeActionChannel)
	return err
}
