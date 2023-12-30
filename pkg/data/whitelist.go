package data

import (
	"github.com/jmoiron/sqlx"
)

type Whitelist struct {
	ID          int64  `db:"id" json:"id"`
	ChannelID   string `db:"channel_id" json:"channel_id"`
	GuildID     string `db:"guild_id" json:"guild_id"`
	ChannelName string `db:"channel_name" json:"channel_name"`
}

type WhitelistModel struct {
	DB *sqlx.DB
}

func (w *WhitelistModel) GetAll() (whitelists []*Whitelist, err error) {
	err = w.DB.Select(&whitelists, "SELECT * FROM whitelist")
	return whitelists, err
}

func (w *WhitelistModel) Insert(whitelist *Whitelist) (err error) {
	_, err = w.DB.Exec(
		"INSERT INTO whitelist (channel_id, guild_id, channel_name) VALUES ($1, $2, $3)",
		whitelist.ChannelID,
		whitelist.GuildID,
		whitelist.ChannelName,
	)
	return err
}

func (w *WhitelistModel) Delete(whitelist *Whitelist) (err error) {
	_, err = w.DB.Exec(
		"DELETE FROM whitelist WHERE channel_id=$1 AND guild_id=$2",
		whitelist.ChannelID,
		whitelist.GuildID,
	)
	return err
}

func (w *WhitelistModel) GetByChannelName(channelName string) (whitelist *Whitelist, err error) {
	err = w.DB.Get(&whitelist, "SELECT * FROM whitelist WHERE channel_name=$1", channelName)
	return whitelist, err
}
