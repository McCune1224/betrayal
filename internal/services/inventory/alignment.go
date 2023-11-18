package inventory

import (
	"github.com/mccune1224/betrayal/internal/util"
)

func (ih *InventoryHandler) SetAlignment(alignment string) error {
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

	return ih.m.Inventories.UpdateProperty(ih.i, "alignment", alignment)
}
