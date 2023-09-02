package data

import (
	"github.com/jmoiron/sqlx"
)

const (
	// Unlimited charges on certain abilities
	unlimited = -1
)

// Ability that is role specific.
type Ability struct {
	ID          int64    `db:"id"`
	Name        string   `db:"name"`
	Description string   `db:"description"`
	Categories  []string `db:"categories"`
	Charges     int      `db:"charges"`
	AnyAbility  bool     `db:"any_ability"`
	// will be listed as 'Role' if AA ability
	Rarity    string `db:"rarity"`
	CreatedAt string `db:"created_at"`
}

type AbilityModel struct {
	DB *sqlx.DB
}
