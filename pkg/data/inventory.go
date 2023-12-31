package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Inventory struct {
	ID             int64          `db:"id" json:"id"`
	DiscordID      string         `db:"discord_id" json:"discord_id"`
	UserPinChannel string         `db:"user_pin_channel" json:"user_pin_channel"`
	UserPinMessage string         `db:"user_pin_message" json:"user_pin_message"`
	RoleName       string         `db:"role_name" json:"role_name"`
	Alignment      string         `db:"alignment" json:"alignment"`
	Abilities      pq.StringArray `db:"abilities" json:"abilities"`
	AnyAbilities   pq.StringArray `db:"any_abilities" json:"any_abilities"`
	Statuses       pq.StringArray `db:"statuses" json:"statuses"`
	Immunities     pq.StringArray `db:"immunities" json:"immunities"`
	Effects        pq.StringArray `db:"effects" json:"effects"`
	Items          pq.StringArray `db:"items" json:"items"`
	ItemLimit      int            `db:"item_limit" json:"item_limit"`
	Perks          pq.StringArray `db:"perks" json:"perks"`
	IsAlive        bool           `db:"is_alive" json:"is_alive"`
	Coins          int64          `db:"coins" json:"coins"`
	CoinBonus      float32        `db:"coin_bonus" json:"coin_bonus"`
	Luck           int64          `db:"luck" json:"luck"`
	Notes          pq.StringArray `db:"notes" json:"notes"`
	CreatedAt      string         `db:"created_at" json:"created_at"`
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

func (m *InventoryModel) UpdateCoinBonus(inventory *Inventory) error {
	query := `UPDATE inventories set coin_bonus=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.CoinBonus, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) UpdateLuck(inventory *Inventory) error {
	query := `UPDATE inventories SET luck=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.Luck, inventory.DiscordID)
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

func (m *InventoryModel) UpdateRoleName(inventory *Inventory) error {
	query := `UPDATE inventories SET role_name=$1 WHERE discord_id=$2`
	_, err := m.DB.Exec(query, inventory.RoleName, inventory.DiscordID)
	if err != nil {
		return err
	}
	return nil
}

func (m *InventoryModel) GetAll() ([]Inventory, error) {
	query := `SELECT * FROM inventories`
	var inventories []Inventory
	err := m.DB.Select(&inventories, query)
	if err != nil {
		return nil, err
	}
	return inventories, nil
}

func (m *InventoryModel) GetAllActiveRoleNames() ([]string, error) {
	query := `SELECT role_name FROM inventories`
	var roleNames []string
	err := m.DB.Select(&roleNames, query)
	if err != nil {
		return nil, err
	}

	return roleNames, nil
}

func (m *InventoryModel) GetAllPlayerIDs() ([]string, error) {
	query := `SELECT discord_id FROM inventories`
	var discordIDs []string
	err := m.DB.Select(&discordIDs, query)
	if err != nil {
		return nil, err
	}

	return discordIDs, nil
}
