package inventory

import (
	"github.com/mccune1224/betrayal/internal/util"
)

func (ih *InventoryHandler) UpdateAlignment(alignment string) error {
	options := []string{"GOOD", "NEUTRAL", "EVIL"}

	// max int value
	low := 1 << 31
	for _, option := range options {
		distance := util.LevenshteinDistance(alignment, option)
		if distance < low {
			low = distance
			alignment = option
		}
	}

	err := ih.m.Inventories.UpdateProperty(ih.i, "alignment", alignment)
	if err != nil {
		return err
	}
	return nil
}
