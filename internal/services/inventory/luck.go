package inventory

import "github.com/mccune1224/betrayal/internal/data"

// Magic number where 100% mythical drop
const maxLuck = 398

func (ih *InventoryHandler) AddLuck(amount int64) error {
	ih.i.Luck += amount
	if ih.i.Luck > maxLuck {
		ih.i.Luck = maxLuck
	}
	return ih.m.Inventories.UpdateLuck(ih.i)
}

func (ih *InventoryHandler) RemoveLuck(amount int64) error {
	ih.i.Luck -= amount
	if ih.i.Luck < 0 {
		ih.i.Luck = 0
	}
	return ih.m.Inventories.UpdateLuck(ih.i)
}

func (ih *InventoryHandler) SetLuck(size int64) error {
	ih.i.Luck = size
	return ih.m.Inventories.UpdateLuck(ih.i)
}

// Determines the luck a player will get from their alliance.
// Good and Evil get +2 per each other member in the same alliance, and +1 for neutral.
func (ih *InventoryHandler) CalculateAllianceLuck(memberIDs []string) (int64, error) {
	var luck int64
	var members []*data.Inventory
	for _, id := range memberIDs {
		if id == ih.i.DiscordID {
			continue
		}
		i, err := ih.m.Inventories.GetByDiscordID(id)
		if err != nil {
			return 0, err
		}
		members = append(members, i)
	}

	pAlignmnet := ih.i.Alignment
	for _, m := range members {
		switch m.Alignment {
		case "GOOD":
			if pAlignmnet == "GOOD" {
				luck += 2
			} else {
				luck += 1
			}
		case "EVIL":
			if pAlignmnet == "EVIL" {
				luck += 2
			} else {
				luck += 1
			}
		case "NEUTRAL":
			luck += 1
		}
	}

	return luck, nil
}
