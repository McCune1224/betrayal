package inventory

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/pkg/data"
)

var ErrAbilityNotFound = errors.New("ability not found")

// InventoryAbility is a type alias for the name of an ability
// Follows the format of "Ability Name [Charges]"
type AbilityString string

func (as *AbilityString) GetName() string {
	return strings.Split(string(*as), " [")[0]
}

func (as *AbilityString) GetCharges() int {
	charge := strings.Split(string(*as), " [")[1]
	charge = strings.TrimSuffix(charge, "]")
	var charges int
	fmt.Sscanf(charge, "%d", &charges)
	return charges
}

// Returns a new InventoryAbility with the given name
func (as *AbilityString) SetName(name string) AbilityString {
	return AbilityString(fmt.Sprintf("%s [%d]", name, as.GetCharges()))
}

// Returns a new InventoryAbility with the given charge
func (aas *AbilityString) SetCharges(charge int) AbilityString {
	return AbilityString(fmt.Sprintf("%s [%d]", aas.GetName(), charge))
}

// Build Inventory Item from Ability
func NewAbilityStringFromAA(ab *data.AnyAbility, chargeOpt ...int) *AbilityString {
	charge := 1
	if len(chargeOpt) > 0 {
		charge = chargeOpt[0]
	}
	newString := fmt.Sprintf("%s [%d]", ab.Name, charge)
	ia := AbilityString(newString)
	return &ia
}

func NewAbilityStringFromAbility(ab *data.Ability, chargeOpt ...int) *AbilityString {
	charge := ab.Charges
	if len(chargeOpt) > 0 {
		charge = chargeOpt[0]
	}
	newString := fmt.Sprintf("%s [%d]", ab.Name, charge)
	ia := AbilityString(newString)
	return &ia
}

func (ih *InventoryHandler) AddAbility(abilityName string, chargeOpt ...int) (AbilityString, error) {
	best, err := ih.m.Abilities.GetByFuzzy(abilityName)
	if err != nil {
		return "", err
	}

	charge := best.Charges
	if len(chargeOpt) > 0 {
		charge = chargeOpt[0]
	}

	for i, ab := range ih.i.Abilities {
		abs := AbilityString(ab)
		if strings.EqualFold(abs.GetName(), best.Name) {
			ih.i.Abilities[i] = string(abs.SetCharges(abs.GetCharges() + charge))
			err := ih.m.Inventories.UpdateAbilities(ih.i)
			if err != nil {
				return "", err
			}
		}
	}

	ih.i.Abilities = append(ih.i.Abilities, string(*NewAbilityStringFromAbility(best, charge)))
	err = ih.m.Inventories.UpdateAbilities(ih.i)
	if err != nil {
		return "", err
	}

	return AbilityString(ih.i.Abilities[len(ih.i.Abilities)-1]), nil
}

func (ih *InventoryHandler) RemoveAbility(abilityName string) (string, error) {
	best, _ := util.FuzzyFind(abilityName, ih.i.Abilities)
	i := slices.Index(ih.i.Abilities, best)
	if i == -1 {
		return "", ErrAbilityNotFound
	}
	removed := AbilityString(ih.i.Abilities[i])
	ih.i.Abilities = append(ih.i.Abilities[:i], ih.i.Abilities[i+1:]...)
	err := ih.m.Inventories.UpdateAbilities(ih.i)
	if err != nil {
		return "", err
	}
	return string(removed), nil
}

func (ih *InventoryHandler) SetAbilityCharges(abilityName string, charge int) error {
	best, _ := util.FuzzyFind(abilityName, ih.i.Abilities)
	bestIaas := AbilityString(best)
	i := slices.Index(ih.i.Abilities, best)
	if i == -1 {
		return ErrAbilityNotFound
	}
	ih.i.AnyAbilities[i] = string(bestIaas.SetCharges(charge))
	return ih.m.Inventories.UpdateAbilities(ih.i)
}
