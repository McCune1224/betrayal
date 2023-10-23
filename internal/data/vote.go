package data

import "github.com/jmoiron/sqlx"

type Vote struct {
	ID        int    `db:"id"`
	ChannelID string `db:"channel_id"`
}

type VoteModel struct {
	DB *sqlx.DB
}

func (v *VoteModel) Insert(vote Vote) error {
	// Delete all entries before hand
	v.DB.Exec("DELETE FROM vote")

	_, err := v.DB.Exec("INSERT INTO vote (channel_id) VALUES ($1)", vote.ChannelID)
	return err
}

func (v *VoteModel) Get() (*Vote, error) {
	var vote Vote
	// Really only one row in the table so just select all and grab the first
	err := v.DB.Get(&vote, "SELECT * FROM vote")
	return &vote, err
}

func (v *VoteModel) Delete(vote Vote) error {
	_, err := v.DB.Exec("DELETE FROM vote WHERE channel_id = $1", vote.ChannelID)
	return err
}
