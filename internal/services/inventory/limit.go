package inventory

func (ih *InventoryHandler) AddLimit(limit int) error {
	ih.i.ItemLimit += limit
	return ih.m.Inventories.UpdateItemLimit(ih.i)
}

func (ih *InventoryHandler) RemoveLimit(limit int) error {
	ih.i.ItemLimit -= limit
	if limit < 0 {
		limit = 0
	}
	return ih.m.Inventories.UpdateItemLimit(ih.i)
}

func (ih *InventoryHandler) SetLimit(limit int) error {
	if limit < 0 {
		limit = 0
	}
	ih.i.ItemLimit = limit
	return ih.m.Inventories.UpdateItemLimit(ih.i)
}
