package data

import "github.com/jmoiron/sqlx"

type Hitlist struct {
	ID         int64  `db:"id"`
	PinChannel string `db:"pin_channel"`
	PinMessage string `db:"pin_message"`
}

type HitlistModel struct {
	DB *sqlx.DB
}

func (hl *HitlistModel) Insert(h *Hitlist) (int64, error) {
	hl.DB.Exec("DELETE FROM hitlist")

	query := `INSERT INTO hitlist (pin_channel, pin_message) VALUES ($1, $2) RETURNING id`
	var id int64
	err := hl.DB.QueryRow(query, h.PinChannel, h.PinMessage).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (hl *HitlistModel) Get() (*Hitlist, error) {
	var h Hitlist
	err := hl.DB.Get(&h, "SELECT * FROM hitlist")
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (hl *HitlistModel) Delete(id int64) error {
	_, err := hl.DB.Exec("DELETE FROM hitlist WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}
