package inventory

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mccune1224/betrayal/internal/util"
)

var (
	ErrAnyAbilityNotFound      = errors.New("ability not found")
	ErrAnyAbilityAlreadyExists = errors.New("ability already exists")
)

// Will insert or update a given any ability for an inventory. Will default charge to 1 if not provided.
func (ih *InventoryHandler) AddAnyAbility(abilityName string, chargeOpt ...int) (AbilityString, error) {
	charge := 1
	if len(chargeOpt) > 0 {
		charge = chargeOpt[0]
	}
	best, err := ih.m.Abilities.GetAnyAbilitybyFuzzy(abilityName)
	if err != nil {
		return "", err
	}
	for i, aa := range ih.i.AnyAbilities {
		aas := AbilityString(aa)
		if strings.EqualFold(aas.GetName(), best.Name) {
			ih.i.AnyAbilities[i] = string(aas.SetCharges(aas.GetCharges() + charge))
			err := ih.m.Inventories.UpdateAnyAbilities(ih.i)
			if err != nil {
				return "", err
			}
			return AbilityString(ih.i.AnyAbilities[i]), nil
		}
	}
	ih.i.AnyAbilities = append(ih.i.AnyAbilities, string(*NewAbilityStringFromAA(best, charge)))
	err = ih.m.Inventories.UpdateAnyAbilities(ih.i)
	if err != nil {
		return "", err
	}
	return AbilityString(ih.i.AnyAbilities[len(ih.i.AnyAbilities)-1]), nil
}

func (ih *InventoryHandler) RemoveAnyAbility(aaName string) (string, error) {
	best, _ := util.FuzzyFind(aaName, ih.i.AnyAbilities)
	i := slices.Index(ih.i.AnyAbilities, best)
	removed := AbilityString(ih.i.AnyAbilities[i])
	ih.i.AnyAbilities = append(ih.i.AnyAbilities[:i], ih.i.AnyAbilities[i+1:]...)
	err := ih.m.Inventories.UpdateAnyAbilities(ih.i)
	if err != nil {
		return "", err
	}
	return string(removed), nil
}

func (ih *InventoryHandler) SetAnyAbilityCharges(aaName string, charge int) error {
	best, _ := util.FuzzyFind(aaName, ih.i.AnyAbilities)
	bestIaas := AbilityString(best)
	i := slices.Index(ih.i.AnyAbilities, best)
	if i == -1 {
		return ErrAnyAbilityNotFound
	}
	ih.i.AnyAbilities[i] = fmt.Sprintf("%s [%d]", bestIaas.GetName(), charge)
	return ih.m.Inventories.UpdateAnyAbilities(ih.i)
}
