package data

import "github.com/jmoiron/sqlx"

type Inventory struct {
	ID              int64  `db:"id"`
	DiscordID       string `db:"discord_id"`
	UserPinChannel  string `db:"user_pin_channel"`
	UserPinMessage  string `db:"user_pin_message"`
	AdminPinChannel string `db:"admin_pin_channel"`
	AdminPinMessage string `db:"admin_pin_message"`
	Content         string `db:"content"`
	CreatedAt       string `db:"created_at"`
}

type InventoryModel struct {
	DB *sqlx.DB
}

func (m *InventoryModel) Insert(inventory *Inventory) (int64, error) {
	query := `
        INSERT INTO inventories (discord_id, user_pin_channel, 
        user_pin_message, admin_pin_channel, admin_pin_message, content)
        VALUES ($1, $2, $3, $4, $5 , $6)
    `
	_, err := m.DB.Exec(query, inventory.DiscordID,
		inventory.UserPinChannel, inventory.UserPinMessage,
		inventory.AdminPinChannel, inventory.AdminPinMessage, inventory.Content)
	if err != nil {
		return -1, err
	}
	var lastInsert Inventory
	err = m.DB.Get(&lastInsert, "SELECT * FROM inventory ORDER BY id DESC LIMIT 1")

	return lastInsert.ID, nil
}

func (m *InventoryModel) GetByDiscordID(discordID string) (*Inventory, error) {
	var inventory Inventory
	err := m.DB.Get(&inventory, "SELECT * FROM inventories WHERE discord_id = $1", discordID)
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (m *InventoryModel) Update(inventory *Inventory) error {
	query := `
        UPDATE inventories SET user_pin_channel = $1, user_pin_message = $2, 
        admin_pin_channel = $3, admin_pin_message = $4, content = $5
        WHERE discord_id = $6
    `
	_, err := m.DB.Exec(
		query,
		inventory.UserPinChannel,
		inventory.UserPinMessage,
		inventory.AdminPinChannel,
		inventory.AdminPinMessage,
		inventory.Content,
		inventory.DiscordID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) Delete(discordID string) error {
	_, err := m.DB.Exec("DELETE FROM inventories WHERE discord_id = $1", discordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) AdminPinUpdate(inventory *Inventory) error {
	query := `
        UPDATE inventories SET admin_pin_channel = $1, 
        admin_pin_message = $2 WHERE discord_id = $3
    `
	_, err := m.DB.Exec(
		query,
		inventory.AdminPinChannel,
		inventory.AdminPinMessage,
		inventory.DiscordID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UserPinUpdate(inventory *Inventory) error {
	query := `
        UPDATE inventories SET user_pin_channel = $1, 
        user_pin_message = $2 WHERE discord_id = $3
    `
	_, err := m.DB.Exec(
		query,
		inventory.UserPinChannel,
		inventory.UserPinMessage,
		inventory.DiscordID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateContent(inventory *Inventory) error {
	query := `
        UPDATE inventories SET content = $1 WHERE discord_id = $2
    `
	_, err := m.DB.Exec(
		query,
		inventory.Content,
		inventory.DiscordID,
	)
	if err != nil {
		return err
	}
	return nil
}
