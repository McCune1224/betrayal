package inventory

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/util"
)

var (
	ErrAnyAbilityNotFound      = errors.New("ability not found")
	ErrAnyAbilityAlreadyExists = errors.New("ability already exists")
)

// InventoryAbility is a type alias for the name of an ability
// Follows the format of "Ability Name [Charges]"
type AnyAbilityString string

func (aas *AnyAbilityString) GetName() string {
	return strings.Split(string(*aas), " [")[0]
}

func (aas *AnyAbilityString) GetCharges() int {
	charge := strings.Split(string(*aas), " [")[1]
	charge = strings.TrimSuffix(charge, "]")
	var charges int
	fmt.Sscanf(charge, "%d", &charges)
	return charges
}

// Returns a new InventoryAbility with the given name
func (aas *AnyAbilityString) SetName(name string) AnyAbilityString {
	return AnyAbilityString(fmt.Sprintf("%s [%d]", name, aas.GetCharges()))
}

// Returns a new InventoryAbility with the given charge
func (aas *AnyAbilityString) SetCharges(charge int) AnyAbilityString {
	return AnyAbilityString(fmt.Sprintf("%s [%d]", aas.GetName(), charge))
}

// Build Inventory Item from Ability
func NewAnyAbilityString(ab *data.AnyAbility, chargeOpt ...int) *AnyAbilityString {
	charge := 1
	if len(chargeOpt) > 0 {
		charge = chargeOpt[0]
	}
	newString := fmt.Sprintf("%s [%d]", ab.Name, charge)
	ia := AnyAbilityString(newString)
	return &ia
}

// Will insert or update a given any ability for an inventory
func (ih *InventoryHandler) AddAnyAbility(abilityName string, chargeOpt ...int) (AnyAbilityString, error) {
	charge := 1
	if len(chargeOpt) > 0 {
		charge = chargeOpt[0]
	}
	best, err := ih.m.Abilities.GetAnyAbilitybyFuzzy(abilityName)
	if err != nil {
		return "", err
	}
	for i, aa := range ih.i.AnyAbilities {
		aas := AnyAbilityString(aa)
		if strings.EqualFold(aas.GetName(), best.Name) {
			ih.i.AnyAbilities[i] = string(aas.SetCharges(aas.GetCharges() + charge))
			err := ih.m.Inventories.UpdateAnyAbilities(ih.i)
			if err != nil {
				return "", err
			}
			return AnyAbilityString(ih.i.AnyAbilities[i]), nil
		}
	}
	ih.i.AnyAbilities = append(ih.i.AnyAbilities, string(*NewAnyAbilityString(best, charge)))
	err = ih.m.Inventories.UpdateAnyAbilities(ih.i)
	if err != nil {
		return "", err
	}
	return AnyAbilityString(ih.i.AnyAbilities[len(ih.i.AnyAbilities)-1]), nil
}

func (ih *InventoryHandler) RemoveAnyAbility(aaName string) (string, error) {
	best, _ := util.FuzzyFind(aaName, ih.i.AnyAbilities)
	i := slices.Index(ih.i.AnyAbilities, best)
	removed := AnyAbilityString(ih.i.AnyAbilities[i])
	ih.i.AnyAbilities = append(ih.i.AnyAbilities[:i], ih.i.AnyAbilities[i+1:]...)
	return string(removed), nil
}

func (ih *InventoryHandler) SetAnyAbilityCharges(aaName string, charge int) error {
	best, _ := util.FuzzyFind(aaName, ih.i.AnyAbilities)
	bestIaas := AnyAbilityString(best)
	i := slices.Index(ih.i.AnyAbilities, best)
	ih.i.AnyAbilities[i] = string(bestIaas.SetCharges(charge))

	return ErrAnyAbilityNotFound
}
