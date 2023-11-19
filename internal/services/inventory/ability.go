package inventory

import (
	"fmt"
	"strings"

	"github.com/mccune1224/betrayal/internal/data"
)

// InventoryAbility is a type alias for the name of an ability
// Follows the format of "Ability Name [Charges]"
type InventoryAbility string

func (ia *InventoryAbility) Name() string {
	return strings.Split(string(*ia), " [")[0]
}

func (ia *InventoryAbility) Charges() int {
	charge := strings.Split(string(*ia), " [")[1]
	charge = strings.TrimSuffix(charge, "]")
	var charges int
	fmt.Sscanf(charge, "%d", &charges)
	return charges
}

// Build Inventory Item from Ability
func NewInventoryAbility(ab *data.Ability) *InventoryAbility {
	newString := fmt.Sprintf("%s [%d]", ab.Name, ab.Charges)
	ia := InventoryAbility(newString)
	return &ia
}
