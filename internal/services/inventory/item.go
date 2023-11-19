package inventory

import (
	"errors"
	"strings"

	"github.com/mccune1224/betrayal/internal/util"
)

var ErrItemNotFound = errors.New("item not found")

func (ih *InventoryHandler) AddItem(name string) (string, error) {
	item, err := ih.m.Items.GetByFuzzy(name)
	if err != nil {
		return "", err
	}

	ih.i.Items = append(ih.i.Items, item.Name)
	err = ih.m.Inventories.UpdateItems(ih.i)
	if err != nil {
		return "", err
	}
	return item.Name, nil
}

func (ih *InventoryHandler) RemoveItem(item string) (string, error) {
	best, _ := util.FuzzyFind(item, ih.i.Items)
	for k, v := range ih.i.Items {
		if strings.EqualFold(v, best) {
			ih.i.Items = append(ih.i.Items[:k], ih.i.Items[k+1:]...)
			return best, ih.m.Inventories.UpdateItems(ih.i)
		}
	}
	return "", ErrItemNotFound
}
