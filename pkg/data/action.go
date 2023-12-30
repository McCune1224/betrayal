package data

import "github.com/jmoiron/sqlx"

type FunnelChannel struct {
	ID         int64  `db:"id" json:"id"`
	ChannelID  string `db:"channel_id" json:"channel_id"`
	GuildID    string `db:"guild_id" json:"guild_id"`
	CurrentDay int    `db:"current_day" json:"current_day"`
}

type FunnelChannelModel struct {
	DB *sqlx.DB
}

func (fcm *FunnelChannelModel) Insert(funnelChannel *FunnelChannel) (int64, error) {
	// Empty out table before inserting as there should only be
	// one funnel channel per guild
	_, err := fcm.DB.Exec("DELETE FROM funnel_channels")
	if err != nil {
		return -1, err
	}

	query := `INSERT INTO funnel_channels ` + PSQLGeneratedInsert(funnelChannel) + ` RETURNING id`
	_, err = fcm.DB.NamedExec(query, &funnelChannel)
	if err != nil {
		return -1, err
	}
	var lastInsert FunnelChannel
	err = fcm.DB.Get(&lastInsert, "SELECT * FROM funnel_channels ORDER BY id DESC LIMIT 1")
	return lastInsert.ID, nil
}

func (fcm *FunnelChannelModel) Remove(guildID string) error {
	_, err := fcm.DB.Exec("DELETE FROM funnel_channels WHERE guild_id = $1", guildID)
	return err
}

func (fcm *FunnelChannelModel) Get(guildID string) (*FunnelChannel, error) {
	var funnelChannel FunnelChannel
	err := fcm.DB.Get(&funnelChannel, "SELECT * FROM funnel_channels WHERE guild_id = $1", guildID)
	if err != nil {
		return nil, err
	}
	return &funnelChannel, nil
}

type Action struct {
	ID                 int64  `db:"id" json:"id"`
	RequestedAction    string `db:"requested_action" json:"requested_action"`
	RequestedChannelID string `db:"requested_channel_id" json:"requested_channel_id"`
	RequestedMessageID string `db:"requested_message_id" json:"requested_message_id"`
	RequesterID        string `db:"requester_id" json:"requester_id"`
	RequestedAt        string `db:"requested_at" json:"requested_at"`
	RequestedDay       int64  `db:"requested_day" json:"requested_day"`
}

type ActionModel struct {
	DB *sqlx.DB
}

func (am *ActionModel) Insert(action *Action) (int64, error) {
	query := `INSERT INTO actions ` + PSQLGeneratedInsert(action) + ` RETURNING id`
	_, err := am.DB.NamedExec(query, &action)
	if err != nil {
		return -1, err
	}
	var lastInsert Action
	err = am.DB.Get(&lastInsert, "SELECT * FROM actions ORDER BY id DESC LIMIT 1")
	return lastInsert.ID, nil
}

func (am *ActionModel) Update(action *Action) error {
	query := `UPDATE actions SET ` + PSQLGeneratedUpdate(action) + ` WHERE id = :id`
	_, err := am.DB.NamedExec(query, &action)
	return err
}

func (am *ActionModel) Get(id int64) (*Action, error) {
	var action Action
	err := am.DB.Get(&action, "SELECT * FROM actions WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &action, nil
}
