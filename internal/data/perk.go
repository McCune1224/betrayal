package data

import "github.com/jmoiron/sqlx"

type Perk struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
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
	err := pm.DB.Get(&p, "SELECT * FROM perks WHERE name ILIKE $1", name)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (pm *PerkModel) Update(p *Perk) error {
	query := `UPDATE perks SET name = $1, description = $2 WHERE id = $3`
	_, err := pm.DB.Exec(query, p.Name, p.Description, p.ID)
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
