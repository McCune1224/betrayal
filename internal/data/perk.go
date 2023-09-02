package data

import "github.com/jmoiron/sqlx"

type Perk struct {
	ID         int64  `db:"id"`
	Name       string `db:"name"`
	Rarity     string `db:"rarity"`
	Categories string `db:"categories"`
	Effect     string `db:"effect"`
	CreatedAt  string `db:"created_at"`
}

type PerkModel struct {
	DB *sqlx.DB
}
