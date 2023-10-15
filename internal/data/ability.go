package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	// Unlimited charges on certain abilities
	unlimited = -1
)

// Ability that is role specific.
type Ability struct {
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	Categories  pq.StringArray `db:"categories"`
	Charges     int            `db:"charges"`
	AnyAbility  bool           `db:"any_ability"`
	// will be listed as 'Role' if AA ability
	Rarity    string `db:"rarity"`
	CreatedAt string `db:"created_at"`
}

type AbilityModel struct {
	DB *sqlx.DB
}

func (am *AbilityModel) Insert(a *Ability) (int64, error) {
	query := `INSERT INTO abilities (name, description, categories, charges, any_ability, rarity)
    VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := am.DB.Exec(
		query,
		a.Name,
		a.Description,
		a.Categories,
		a.Charges,
		a.AnyAbility,
		a.Rarity,
	)
	if err != nil {
		return -1, err
	}
	var lastInsert Ability
	err = am.DB.Get(&lastInsert, "SELECT * FROM abilities ORDER BY id DESC LIMIT 1")

	if err != nil {
		return -1, err
	}

	return lastInsert.ID, nil
}

func (am *AbilityModel) Get(id int64) (*Ability, error) {
	var a Ability
	err := am.DB.Get(&a, "SELECT * FROM Abilities WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (am *AbilityModel) GetByName(name string) (*Ability, error) {
	var a Ability
	err := am.DB.Get(&a, "SELECT * FROM Abilities WHERE name ILIKE $1", name)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (am *AbilityModel) GetAll() ([]Ability, error) {
	var abilities []Ability
	err := am.DB.Select(&abilities, "SELECT * FROM Abilities")
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func (am *AbilityModel) GetByCategory(category string) ([]Ability, error) {
	var abilities []Ability
	err := am.DB.Select(&abilities, "SELECT * FROM Abilities WHERE categories ILIKE $1", category)
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func (am *AbilityModel) GetByRarity(rarity string) ([]Ability, error) {
	var abilities []Ability
	err := am.DB.Select(&abilities, "SELECT * FROM Abilities WHERE rarity ILIKE $1", rarity)
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func (am *AbilityModel) GetRandomByRarity(rarity string) (*Ability, error) {
	var a Ability
	err := am.DB.Get(
		&a,
		"SELECT * FROM Abilities WHERE rarity ILIKE $1 ORDER BY RANDOM() LIMIT 1",
		rarity,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (am *AbilityModel) Update(a *Ability) error {
	query := `UPDATE Abilities SET name = $1, description = $2, categories = $3, charges = $4, any_ability = $5, rarity = $6 WHERE id = $7`
	_, err := am.DB.Exec(
		query,
		a.Name,
		a.Description,
		a.Categories,
		a.Charges,
		a.AnyAbility,
		a.Rarity,
		a.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (am *AbilityModel) Delete(id int64) error {
	_, err := am.DB.Exec("DELETE FROM Abilities WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (am *AbilityModel) WipeTable() error {
	_, err := am.DB.Exec("DELETE FROM Abilities")
	if err != nil {
		return err
	}
	return nil
}

func (am *AbilityModel) Upsert(a *Ability) error {
	query := `INSERT INTO Abilities (name, description, categories, charges, any_ability, rarity)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (name) DO UPDATE
    SET name = $1, description = $2, categories = $3, charges = $4, any_ability = $5, rarity = $6`
	_, err := am.DB.Exec(
		query,
		a.Name,
		a.Description,
		a.Categories,
		a.Charges,
		a.AnyAbility,
		a.Rarity,
	)
	if err != nil {
		return err
	}
	return nil
}
