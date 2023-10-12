package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Inventory struct {
	ID             int64          `db:"id"`
	DiscordID      string         `db:"discord_id"`
	UserPinChannel string         `db:"user_pin_channel"`
	UserPinMessage string         `db:"user_pin_message"`
	RoleName       string         `db:"role_name"`
	Alignment      string         `db:"alignment"`
	Abilities      pq.StringArray `db:"abilities"`
	AnyAbilities   pq.StringArray `db:"any_abilities"`
	Statuses       pq.StringArray `db:"statuses"`
	Immunities     pq.StringArray `db:"immunities"`
	Effects        pq.StringArray `db:"effects"`
	Items          pq.StringArray `db:"items"`
	ItemLimit      int            `db:"item_limit"`
	Perks          pq.StringArray `db:"perks"`
	Coins          int64          `db:"coins"`
	CoinBonus      float32        `db:"coin_bonus"`
	Notes          pq.StringArray `db:"notes"`
	CreatedAt      string         `db:"created_at"`
}

type InventoryModel struct {
	DB *sqlx.DB
}

func (m *InventoryModel) Insert(i *Inventory) (int64, error) {
	query := `INSERT INTO inventories ` + PSQLGeneratedInsert(i) + ` RETURNING id`
	_, err := m.DB.NamedExec(query, &i)
	if err != nil {
		return -1, err
	}
	var lastInsert Inventory
	err = m.DB.Get(&lastInsert, "SELECT * FROM inventories ORDER BY id DESC LIMIT 1")
	return lastInsert.ID, nil
}

func (m *InventoryModel) GetByDiscordID(discordID string) (*Inventory, error) {
	query := `SELECT * FROM inventories WHERE discord_id=$1`
	var inventory Inventory
	err := m.DB.Get(&inventory, query, discordID)
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (m *InventoryModel) GetByPinChannel(pinChannel string) (*Inventory, error) {
	query := `SELECT * FROM inventories WHERE user_pin_channel=$1`
	var inventory Inventory
	err := m.DB.Get(&inventory, query, pinChannel)
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (m *InventoryModel) Update(inventory *Inventory) error {
	query := `UPDATE inventories SET ` + PSQLGeneratedUpdate(
		inventory,
	) + ` WHERE discord_id=:discord_id`
	_, err := m.DB.NamedExec(query, &inventory)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateProperty(
	inventory *Inventory,
	columnName string,
	value interface{},
) error {
	query := `UPDATE inventories SET ` + columnName + `=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, value, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateCoins(inventory *Inventory) error {
	query := `UPDATE inventories SET coins=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Coins, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateAbilities(inventory *Inventory) error {
	query := `UPDATE inventories SET abilities=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Abilities, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

// Overwrites the entire abilities column with the new abilities
func (m *InventoryModel) UpdateAnyAbilities(inventory *Inventory) error {
	query := `UPDATE inventories SET any_abilities=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.AnyAbilities, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdatePerks(inventory *Inventory) error {
	query := `UPDATE inventories SET perks=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Perks, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateItems(inventory *Inventory) error {
	query := `UPDATE inventories SET items=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Items, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateStatuses(inventory *Inventory) error {
	query := `UPDATE inventories SET statuses=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Statuses, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateImmunities(inventory *Inventory) error {
	query := `UPDATE inventories SET immunities=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Immunities, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateEffects(inventory *Inventory) error {
	query := `UPDATE inventories SET effects=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Effects, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateNotes(inventory *Inventory) error {
	query := `UPDATE inventories SET notes=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Notes, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) Delete(discordID string) error {
	query := `DELETE FROM inventories WHERE discord_id=$1`
	_, err := m.DB.Exec(query, discordID)
	if err != nil {
		return err
	}
	return nil

}

func (m *InventoryModel) UpdateItemLimit(inventory *Inventory) error {
	query := `UPDATE inventories SET item_limit=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.ItemLimit, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}
