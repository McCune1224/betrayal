package inventory

import (
	"strings"

	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/pkg/data"
)

// Handles all inventory related actions
type InventoryHandler struct {
	m data.Models
	i *data.Inventory
}

// Create a new InventoryHandler whthout an Inventory (normally used for creating a new inventory)
func InitInventoryHandler(models data.Models, inv ...*data.Inventory) *InventoryHandler {
	ih := &InventoryHandler{
		m: models,
	}

	if len(inv) > 0 {
		ih.i = inv[0]
	}
	return ih
}

func (ih *InventoryHandler) GetInventory() *data.Inventory {
	return ih.i
}

func (ih *InventoryHandler) RefreshInventory() error {
	inv, err := ih.m.Inventories.GetByDiscordID(ih.i.DiscordID)
	if err != nil {
		return err
	}
	ih.i = inv
	return nil
}

func (ih *InventoryHandler) CreateInventory(initInv *data.Inventory) error {
	inv := initInv
	// FIXME: Lord please forgive for the unholy amount of switch statements I am about to unleash
	// Will need to make some sort of Website or UI to allow for custom roles to be created instead of me hardcoding them
	roleName := strings.ToLower(inv.RoleName)
	switch roleName {
	// --- GOOD ROLES ---
	case "cerberus":
		// Due to perk Hades' Hound
		inv.Immunities = pq.StringArray{"Frozen", "Burned"}
	case "detective":
		// Due to perk Clever
		inv.Immunities = pq.StringArray{"Blackmailed", "Disabled", "Despaired"}
	case "fisherman":
		// Due to perk Barrels
		inv.ItemLimit = 8
	case "hero":
		// Due to perk Compos Mentis
		inv.Immunities = pq.StringArray{"Madness"}
	case "nurse":
		// Due to perk Powerful Immunity
		inv.Immunities = pq.StringArray{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}
	case "terminal":
		// Due to perk Heartbeats
		inv.Immunities = pq.StringArray{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}
	case "wizard":
		// due to perk Magic Barrier
		inv.Immunities = pq.StringArray{"Frozen", "Paralyzed", "Burned", "Cursed"}
	case "yeti":
		// Due to perk Winter Coat
		inv.Immunities = pq.StringArray{"Frozen"}

		// Neutral Roles
	case "cyborg":
		inv.Immunities = pq.StringArray{"Paralyzed", "Frozen", "Burned", "Despaired", "Blackmailed", "Drunk"}
	case "entertainer":
		// Due to perk Top-Hat Tip
		inv.Immunities = pq.StringArray{"Unlucky"}
		inv.Statuses = pq.StringArray{"Lucky"}
	case "magician":
		// Due to perk Top-Hat Tip
		inv.Statuses = pq.StringArray{"Lucky"}
		inv.Immunities = pq.StringArray{"Unlucky"}
	case "masochist":
		// Due to perk One Track Mind
		inv.Immunities = pq.StringArray{"Lucky"}
	case "succubus":
		// Due to perk Dominatrix
		inv.Immunities = pq.StringArray{"Blackmail"}
	//
	// Evil Roles
	case "arsonist":
		// Due to perk Ashes to Ashes / Flamed
		inv.Immunities = pq.StringArray{"Burned"}
	case "cultist":
		inv.Immunities = pq.StringArray{"Curse"}
	case "director":
		inv.Immunities = pq.StringArray{"Despaired", "Blackmailed", "Drunk"}
	case "gatekeeper":
		inv.Immunities = pq.StringArray{"Restrained", "Paralyzed", "Frozen"}
	case "hacker":
		inv.Immunities = pq.StringArray{"Disabled", "Blackmailed"}
	case "highwayman":
		inv.Immunities = pq.StringArray{"Madness"}
	case "imp":
		inv.Immunities = pq.StringArray{"Despaired", "Paralyzed"}
	case "threatener":
		inv.ItemLimit = 6
	}
	_, err := ih.m.Inventories.Insert(inv)
	if err != nil {
		return err
	}
	ih.i = inv

	return nil
}
