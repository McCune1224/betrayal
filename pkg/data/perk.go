package data

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mccune1224/betrayal/internal/util"
)

type Perk struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	CreatedAt   string `db:"created_at" json:"created_at"`
}

type PerkModel struct {
	DB *sqlx.DB
}

func (pm *PerkModel) Insert(p *Perk) (int64, error) {
	query := `INSERT INTO perks (name, description) VALUES ($1, $2)`
	_, err := pm.DB.Exec(query, p.Name, p.Description)
	if err != nil {
		return -1, err
	}
	var lastInert Perk
	err = pm.DB.Get(&lastInert, "SELECT * FROM perks ORDER BY id DESC LIMIT 1")

	return lastInert.ID, nil
}

func (pm *PerkModel) Get(id int64) (*Perk, error) {
	var p Perk
	err := pm.DB.Get(&p, "SELECT * FROM perks WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (pm *PerkModel) GetByName(name string) (*Perk, error) {
	var p Perk
	// Fuzzy search for perk
	query := `SELECT * FROM perks WHERE name ILIKE '%' || $1 || '%'`
	err := pm.DB.Get(&p, query, name)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (pm *PerkModel) GetByFuzzy(name string) (*Perk, error) {
	var p Perk
	query := "SELECT * FROM perks WHERE name ILIKE $1"
	pm.DB.Get(&p, query, name)
	if p.Name != "" {
		return &p, nil
	}

	var pChoices []Perk
	if len(name) > 2 {
		return nil, errors.New("search term must be at least 2 characters")
	}

	err := pm.DB.Select(&pChoices, "SELECT * FROM perks")
	if err != nil {
		return nil, err
	}

	strChoices := make([]string, len(pChoices))
	for i, a := range pChoices {
		strChoices[i] = a.Name
	}
	best, _ := util.FuzzyFind(name, strChoices)
	for _, a := range pChoices {
		if a.Name == best {
			p = a
		}
	}
	return &p, nil
}

func (pm *PerkModel) GetAll() ([]Perk, error) {
	var perks []Perk
	err := pm.DB.Select(&perks, "SELECT * FROM perks")
	if err != nil {
		return nil, err
	}
	return perks, nil
}

func (pm *PerkModel) Update(p *Perk) error {
	query := `UPDATE perks SET name = $1, description = $2 WHERE id = $3`
	_, err := pm.DB.Exec(query, p.Name, p.Description, p.ID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PerkModel) UpdateName(perk *Perk) error {
	query := `UPDATE perks SET name = $1 WHERE id = $2`
	_, err := p.DB.Exec(query, perk.Name, perk.ID)
	if err != nil {
		return err
	}
	return nil
}

func (pm *PerkModel) Delete(id int64) error {
	_, err := pm.DB.Exec("DELETE FROM perks WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (pm *PerkModel) Upsert(p *Perk) error {
	query := `INSERT INTO perks (name, description) VALUES ($1, $2) 
    ON CONFLICT (name) DO UPDATE SET name = $1, description = $2`
	_, err := pm.DB.Exec(query, p.Name, p.Description)
	if err != nil {
		return err
	}
	return nil
}
