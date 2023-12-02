package inventory

import (
	"errors"
	"strings"

	"github.com/mccune1224/betrayal/internal/util"
)

var (
	ErrImmunityExists   = errors.New("immunity already exists")
	ErrImmunityNotFound = errors.New("immunity not found")
)

// Add an immunity to inventory and persist to database, if it doesn't already exist
func (ih *InventoryHandler) AddImmunity(immunity string) (string, error) {
	for _, im := range ih.i.Immunities {
		if strings.EqualFold(im, immunity) {
			return "", ErrImmunityExists
		}
	}
	dbImmunity, err := ih.m.Statuses.GetByFuzzy(immunity)
	if err != nil {
		return "", ErrImmunityNotFound
	}
	ih.i.Immunities = append(ih.i.Immunities, dbImmunity.Name)
	return dbImmunity.Name, ih.m.Inventories.UpdateImmunities(ih.i)
}

// Remove an immunity from inventory and persist to database
func (ih *InventoryHandler) RemoveImmunity(immunity string) (string, error) {
	best, _ := util.FuzzyFind(immunity, ih.i.Immunities)
	for i, im := range ih.i.Immunities {
		if strings.EqualFold(im, best) {
			ih.i.Immunities = append(ih.i.Immunities[:i], ih.i.Immunities[i+1:]...)
			return best, ih.m.Inventories.UpdateImmunities(ih.i)
		}
	}
	return "", ErrImmunityNotFound
}
