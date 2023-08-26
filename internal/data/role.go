package data

import "github.com/jmoiron/sqlx"

// General representation of a role in the game in db.
type Role struct {
	Name        string
	Description string
	Alignment   string
	IsActive    bool
	Abilities   []Ability
	Perks       []Perk
}

type RoleModel struct {
	DB *sqlx.DB
}
