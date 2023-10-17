package data

import "github.com/jmoiron/sqlx"

// DB representation of a duel within Betrayal
type Duel struct {
	ID                  int64  `db:"id"`
	BlackPlayerID       string `db:"black_player_id"`
	WhitePlayerID       string `db:"white_player_id"`
	AvailableBlackTiles []int  `db:"available_black_tiles"`
	AvailableWhiteTiles []int  `db:"available_white_tiles"`
	BlackPoints         int    `db:"black_points"`
	WhitePoints         int    `db:"white_points"`
	LeadingPlayer       string `db:"leading_player"`
	RoundNumber         int    `db:"round_number"`
	Winner              int    `db:"winner"`
	IsActive            bool   `db:"is_active"`
	CreatedAt           string `db:"created_at"`
}

type DuelModel struct {
	DB *sqlx.DB
}

func (dm *DuelModel) Insert(duel *Duel) (int64, error) {
	query := `INSERT INTO duels ` + PSQLGeneratedInsert(duel) + ` RETURNING id`
	_, err := dm.DB.NamedExec(query, &duel)
	if err != nil {
		return -1, err
	}
	var lastInsert Duel
	err = dm.DB.Get(&lastInsert, "SELECT * FROM duels ORDER BY id DESC LIMIT 1")
	if err != nil {
		return -1, err
	}
	return lastInsert.ID, nil
}

func (dm *DuelModel) Get(id int64) (*Duel, error) {
	var duel Duel
	err := dm.DB.Get(&duel, "SELECT * FROM duels WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &duel, nil
}

func (dm *DuelModel) UpdateWinner(d *Duel, winner int) error {
	_, err := dm.DB.Exec("UPDATE duels SET winner = $1 WHERE id = $2", winner, d.ID)
	return err
}

func (dm *DuelModel) UpdateRound(d *Duel, round int) error {
	_, err := dm.DB.Exec("UPDATE duels SET round_number = $1 WHERE id = $2", round, d.ID)
	return err
}

func (dm *DuelModel) UpdateLeadingPlayer(d *Duel, turn string) error {
	_, err := dm.DB.Exec("UPDATE duels SET leading_player = $1 WHERE id = $2", turn, d.ID)
	return err
}

func (dm *DuelModel) UpdateBlackPoints(d *Duel, points int) error {
	_, err := dm.DB.Exec("UPDATE duels SET black_points = $1 WHERE id = $2", points, d.ID)
	return err
}

func (dm *DuelModel) UpdateWhitePoints(d *Duel, points int) error {
	_, err := dm.DB.Exec("UPDATE duels SET white_points = $1 WHERE id = $2", points, d.ID)
	return err
}

func (dm *DuelModel) UpdateBlackTiles(d *Duel, newTiles []int) error {
	_, err := dm.DB.Exec("UPDATE duels SET available_black_tiles = $1 WHERE id = $2", newTiles, d.ID)
	return err
}

func (dm *DuelModel) UpdateWhiteTiles(d *Duel, newTiles []int) error {
	_, err := dm.DB.Exec("UPDATE duels SET available_white_tiles = $1 WHERE id = $2", newTiles, d.ID)
	return err
}

func (dm *DuelModel) UpdateIsActive(d *Duel, isActive bool) error {
	_, err := dm.DB.Exec("UPDATE duels SET is_active = $1 WHERE id = $2", isActive, d.ID)
	return err
}

func (dm *DuelModel) UpdateProperty(d *Duel, propertyName string, value interface{}) error {
	query := `UPDATE duels SET ` + propertyName + `=$1 WHERE id=$2`
	_, err := dm.DB.Exec(query, value, d.ID)
	if err != nil {
		return err
	}
	return nil
}

func (dm *DuelModel) Delete(id int64) error {
	_, err := dm.DB.Exec("DELETE FROM duels WHERE id = $1", id)
	return err
}

func (dm *DuelModel) GetActivePlayerDuels(id string) ([]*Duel, error) {
	var duels []*Duel
	err := dm.DB.Select(&duels, "SELECT * FROM duels WHERE (black_player_id = $1 OR white_player_id = $1) AND is_active = true", id)
	if err != nil {
		return nil, err
	}
	return duels, nil
}
