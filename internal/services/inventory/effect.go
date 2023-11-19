package inventory

import (
	"errors"
	"strings"
)

var (
	ErrAlreadyExists  = errors.New("effect already exists")
	ErrEffectNotFound = errors.New("effect not found")
)

func (ih *InventoryHandler) AddEffect(effect string) (string, error) {
	for _, v := range ih.i.Effects {
		if strings.EqualFold(v, effect) {
			return "", ErrAlreadyExists
		}
	}
	ih.i.Effects = append(ih.i.Effects, effect)
	return effect, ih.m.Inventories.UpdateEffects(ih.i)
}

func (ih *InventoryHandler) RemoveEffect(effect string) (string, error) {
	for k, v := range ih.i.Effects {
		if strings.EqualFold(v, effect) {
			ih.i.Effects = append(ih.i.Effects[:k], ih.i.Effects[k+1:]...)
			return effect, ih.m.Inventories.UpdateEffects(ih.i)
		}
	}
	return "", ErrEffectNotFound
}
