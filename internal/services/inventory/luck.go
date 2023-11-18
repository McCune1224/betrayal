package inventory

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
