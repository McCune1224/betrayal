package data

import "github.com/jmoiron/sqlx"

type Player struct {
	ID        int64  `db:"id"`
	DiscordID string `db:"discord_id"`
	RoleID    int64  `db:"role_id"`
	Coins     int  `db:"coins"`
	CreatedAt string `db:"created_at"`
}

type PlayerModel struct {
	DB *sqlx.DB
}

func (pm *PlayerModel) Insert(p *Player) (int64, error) {
	query := `INSERT INTO players (discord_id, role_id, coins) VALUES ($1, $2, $3)`
	_, err := pm.DB.Exec(query, p.DiscordID, p.RoleID, p.Coins)
	if err != nil {
		return -1, err
	}
	var lastInsert Player
	err = pm.DB.Get(&lastInsert, "SELECT * FROM players ORDER BY id DESC LIMIT 1")

	return int64(lastInsert.ID), nil
}

func (pm *PlayerModel) Get(id int64) (*Player, error) {
	var p Player
	err := pm.DB.Get(&p, "SELECT * FROM players WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (pm *PlayerModel) GetByDiscordID(discordID string) (*Player, error) {
	var p Player
	err := pm.DB.Get(&p, "SELECT * FROM players WHERE discord_id = $1", discordID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (pm *PlayerModel) Update(p *Player) error {
	query := `UPDATE players SET role_id = $1, coins = $2 WHERE id = $3`
	_, err := pm.DB.Exec(query, p.RoleID, p.Coins, p.ID)
	if err != nil {
		return err
	}
	return nil
}

func (pm *PlayerModel) Delete(id int64) error {
	_, err := pm.DB.Exec("DELETE FROM players WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (pm *PlayerModel) WipeTable() error {
	_, err := pm.DB.Exec("DELETE FROM players")
	if err != nil {
		return err
	}
	return nil
}
