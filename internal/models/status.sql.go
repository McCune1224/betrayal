// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: status.sql

package models

import (
	"context"
)

const createStatus = `-- name: CreateStatus :one
INSERT INTO status (name, description) VALUES ($1, $2) RETURNING id, name, description
`

type CreateStatusParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (q *Queries) CreateStatus(ctx context.Context, arg CreateStatusParams) (Status, error) {
	row := q.db.QueryRow(ctx, createStatus, arg.Name, arg.Description)
	var i Status
	err := row.Scan(&i.ID, &i.Name, &i.Description)
	return i, err
}

const deleteStatus = `-- name: DeleteStatus :exec
delete from status
where id = $1
`

func (q *Queries) DeleteStatus(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteStatus, id)
	return err
}

const getStatusByFuzzy = `-- name: GetStatusByFuzzy :one
select id, name, description
from status
order by levenshtein(name, $1::varchar(255)) asc
limit 1
`

func (q *Queries) GetStatusByFuzzy(ctx context.Context, dollar_1 string) (Status, error) {
	row := q.db.QueryRow(ctx, getStatusByFuzzy, dollar_1)
	var i Status
	err := row.Scan(&i.ID, &i.Name, &i.Description)
	return i, err
}

const getStatusByName = `-- name: GetStatusByName :one
select id, name, description
from status
where name = $1
`

func (q *Queries) GetStatusByName(ctx context.Context, name string) (Status, error) {
	row := q.db.QueryRow(ctx, getStatusByName, name)
	var i Status
	err := row.Scan(&i.ID, &i.Name, &i.Description)
	return i, err
}

const listStatus = `-- name: ListStatus :many
select id, name, description
from status
`

func (q *Queries) ListStatus(ctx context.Context) ([]Status, error) {
	rows, err := q.db.Query(ctx, listStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Status
	for rows.Next() {
		var i Status
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
