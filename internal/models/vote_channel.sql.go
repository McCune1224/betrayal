// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: vote_channel.sql

package models

import (
	"context"
)

const getVoteChannel = `-- name: GetVoteChannel :one
select channel_id
from vote_channel
limit 1
`

func (q *Queries) GetVoteChannel(ctx context.Context) (string, error) {
	row := q.db.QueryRow(ctx, getVoteChannel)
	var channel_id string
	err := row.Scan(&channel_id)
	return channel_id, err
}

const upsertVoteChannel = `-- name: UpsertVoteChannel :exec
insert into vote_channel (channel_id) values ($1)
returning channel_id
`

func (q *Queries) UpsertVoteChannel(ctx context.Context, channelID string) error {
	_, err := q.db.Exec(ctx, upsertVoteChannel, channelID)
	return err
}

const wipeVoteChannel = `-- name: WipeVoteChannel :exec
delete from vote_channel
`

func (q *Queries) WipeVoteChannel(ctx context.Context) error {
	_, err := q.db.Exec(ctx, wipeVoteChannel)
	return err
}
