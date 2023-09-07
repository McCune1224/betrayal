package data

import (
	"github.com/jmoiron/sqlx"
)

type Status struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CratedAt    string `db:"created_at"`
}

type StatusModel struct {
	DB *sqlx.DB
}
