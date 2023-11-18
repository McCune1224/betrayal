package inventory

func (ih *InventoryHandler) AddNote(note string) error {
	ih.i.Notes = append(ih.i.Notes, note)
	err := ih.m.Inventories.UpdateNotes(ih.i)
	if err != nil {
		return err
	}
	return nil
}

func (ih *InventoryHandler) RemoveNote(index int) error {
	ih.i.Notes = append(ih.i.Notes[:index], ih.i.Notes[index+1:]...)
	err := ih.m.Inventories.UpdateNotes(ih.i)
	if err != nil {
		return err
	}
	return nil
}
