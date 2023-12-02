package inventory

import (
	"errors"
	"log"
	"slices"
	"strings"

	"github.com/mccune1224/betrayal/internal/util"
)

var (
	ErrStatusAlreadyExists = errors.New("status already exists")
	ErrStatusNotFound      = errors.New("status not found")
)

func (ih *InventoryHandler) AddStatus(name string) (string, error) {
	best, err := ih.m.Statuses.GetByFuzzy(name)
	if err != nil {
		return "", err
	}
	for _, s := range ih.i.Statuses {
		if strings.EqualFold(best.Name, s) {
			return "", ErrStatusAlreadyExists
		}
	}
	ih.i.Statuses = append(ih.i.Statuses, best.Name)
	err = ih.m.Inventories.UpdateStatuses(ih.i)

	if err != nil {
		return "", err
	}
	return best.Name, nil
}

func (ih *InventoryHandler) RemoveStatus(name string) (string, error) {
	best, _ := util.FuzzyFind(name, ih.i.Statuses)
	i := slices.Index(ih.i.Statuses, best)
	if i == -1 {
		return "", ErrStatusNotFound
	}
	removed := ih.i.Statuses[i]
	ih.i.Statuses = append(ih.i.Statuses[:i], ih.i.Statuses[i+1:]...)
	err := ih.m.Inventories.UpdateStatuses(ih.i)
	if err != nil {
		log.Println(err)
		return "", err
	}
  return removed, nil
}
