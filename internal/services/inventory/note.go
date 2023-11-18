package inventory

func (ih *InventoryHandler) AddNote(note string) error {
	ih.i.Notes = append(ih.i.Notes, note)
	return ih.m.Inventories.UpdateNotes(ih.i)
}

func (ih *InventoryHandler) RemoveNote(index int) error {
	ih.i.Notes = append(ih.i.Notes[:index], ih.i.Notes[index+1:]...)
	return ih.m.Inventories.UpdateNotes(ih.i)
}
