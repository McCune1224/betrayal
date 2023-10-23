package data

import (
	"github.com/jmoiron/sqlx"
)

type Status struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CratedAt    string `db:"created_at"`
}

type StatusModel struct {
	DB *sqlx.DB
}

func (sm *StatusModel) Insert(s *Status) (int64, error) {
	query := `INSERT INTO statuses (name, description) VALUES ($1, $2)`
	_, err := sm.DB.Exec(query, s.Name, s.Description)
	if err != nil {
		return -1, err
	}
	var lastInsert Status
	err = sm.DB.Get(&lastInsert, "SELECT * FROM statuses ORDER BY id DESC LIMIT 1")
	if err != nil {
		return -1, err
	}

	return lastInsert.ID, nil
}

func (sm *StatusModel) Get(id int64) (*Status, error) {
	var status Status
	err := sm.DB.Get(&status, "SELECT * FROM statuses WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (sm *StatusModel) GetByName(name string) (*Status, error) {
	var status Status

	err := sm.DB.Get(&status, "SELECT * FROM statuses WHERE name ILIKE $1", name)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (sm *StatusModel) GetAll() ([]Status, error) {
	var statuses []Status
	err := sm.DB.Select(&statuses, "SELECT * FROM statuses")
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (sm *StatusModel) Update(status *Status) error {
	query := `UPDATE statuses SET name = $1, description = $2 WHERE id = $3`
	_, err := sm.DB.Exec(query, status.Name, status.Description, status.ID)
	if err != nil {
		return err
	}
	return nil
}

func (sm *StatusModel) Delete(id int64) error {
	_, err := sm.DB.Exec("DELETE FROM statuses WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (sm *StatusModel) Upsert(s *Status) error {
	query := `INSERT INTO statuses (name, description) VALUES ($1, $2)
    ON CONFLICT (name) DO UPDATE SET name = $1, description = $2`
	_, err := sm.DB.Exec(query, s.Name, s.Description)
	if err != nil {
		return err
	}
	return nil
}
