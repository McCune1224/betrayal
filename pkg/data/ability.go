package data

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/util"
)

const (
	// Unlimited charges on certain abilities
	unlimited = -1
)

// Ability that is role specific.
type Ability struct {
	ID          int64          `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	Description string         `db:"description" json:"description"`
	Categories  pq.StringArray `db:"categories" json:"categories"`
	Charges     int            `db:"charges" json:"charges"`
	AnyAbility  bool           `db:"any_ability" json:"any_ability"`
	Rarity      string         `db:"rarity" json:"rarity"`
	CreatedAt   string         `db:"created_at" json:"created_at"`
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
	query := "SELECT * FROM Abilities WHERE name ILIKE $1"
	err := am.DB.Get(&a, query, name)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (am *AbilityModel) GetByFuzzy(name string) (*Ability, error) {
	// attempt get by name first before following here
	var ab Ability
	query := "SELECT * FROM Abilities WHERE name ILIKE $1"
	am.DB.Get(&ab, query, name)
	if ab.Name != "" {
		return &ab, nil
	}

	var abChoices []Ability
	if len(name) < 2 {
		return nil, errors.New("search term must be at least 2 characters")
	}
	err := am.DB.Select(&abChoices, "SELECT * FROM abilities")
	if err != nil {
		return nil, err
	}

	strChoices := make([]string, len(abChoices))
	for i, a := range abChoices {
		strChoices[i] = a.Name
	}
	best, _ := util.FuzzyFind(name, strChoices)
	for _, a := range abChoices {
		if a.Name == best {
			ab = a
		}
	}
	return &ab, nil
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
