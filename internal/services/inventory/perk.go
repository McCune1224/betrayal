package inventory

import (
	"errors"
	"log"
	"slices"
	"strings"

	"github.com/mccune1224/betrayal/internal/util"
)

var (
	ErrPerkAlreadyExists = errors.New("perk already exists")
	ErrPerkNotFound      = errors.New("perk not found")
)

func (ih *InventoryHandler) AddPerk(perk string) (string, error) {
	best, err := ih.m.Perks.GetByFuzzy(perk)
	if err != nil {
		return "", err
	}
	for _, p := range ih.i.Perks {
		if strings.EqualFold(best.Name, p) {
			return "", ErrPerkAlreadyExists
		}
	}
	ih.i.Perks = append(ih.i.Perks, perk)
	err = ih.m.Inventories.UpdatePerks(ih.i)
	if err != nil {
		return "", err
	}

	return best.Name, nil
}

func (ih *InventoryHandler) RemovePerk(perk string) (string, error) {
	best, _ := util.FuzzyFind(perk, ih.i.Perks)
	i := slices.Index(ih.i.Perks, best)
	if i == -1 {
		return "", ErrPerkNotFound
	}
	removed := ih.i.Perks[i]
	ih.i.Perks = append(ih.i.Perks[:i], ih.i.Perks[i+1:]...)
	err := ih.m.Inventories.UpdatePerks(ih.i)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return removed, nil
}
