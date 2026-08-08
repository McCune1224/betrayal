package inventory

import (
	"errors"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddAbility(abilityName string, quantity int32) (*models.AbilityInfo, error) {
	if quantity < 0 {
		return nil, errors.New("quantity must not be negative")
	}
	ctx, cancel := dbCtx()
	defer cancel()
	query := models.New(ih.pool)
	ability, err := query.GetAbilityInfoByFuzzy(ctx, abilityName)
	if err != nil {
		return nil, err
	}
	currentAbilityIds, err := query.ListPlayerAbilityJoin(ctx, ih.player.ID)
	if err != nil {
		return nil, err
	}

	for _, abilityId := range currentAbilityIds {
		if ability.ID == abilityId.AbilityID {
			return nil, errors.New("ability already added")
		}
	}

	if quantity == 0 {
		quantity = ability.DefaultCharges
	}

	_, err = query.CreatePlayerAbilityJoin(ctx, models.CreatePlayerAbilityJoinParams{
		PlayerID:  ih.player.ID,
		AbilityID: ability.ID,
		Quantity:  quantity,
	})
	if err != nil {
		return nil, err
	}

	return &ability, nil
}

func (ih *InventoryHandler) RemoveAbility(abilityName string) (*models.AbilityInfo, error) {
	ctx, cancel := dbCtx()
	defer cancel()
	query := models.New(ih.pool)
	ability, err := query.GetAbilityInfoByFuzzy(ctx, abilityName)
	if err != nil {
		return nil, err
	}
	err = query.DeletePlayerAbility(ctx, models.DeletePlayerAbilityParams{
		PlayerID:  ih.player.ID,
		AbilityID: ability.ID,
	})
	return &ability, err
}

func (ih *InventoryHandler) UpdateAbility(abilityName string, quantity int32) (*models.AbilityInfo, error) {
	if quantity < 0 {
		return nil, errors.New("quantity must not be negative")
	}
	ctx, cancel := dbCtx()
	defer cancel()
	query := models.New(ih.pool)
	ability, err := query.GetAbilityInfoByFuzzy(ctx, abilityName)
	if err != nil {
		return nil, err
	}

	currentAbilityList, err := query.ListPlayerAbilityJoin(ctx, ih.player.ID)
	if err != nil {
		return nil, err
	}
	var targetAbility *models.PlayerAbility
	for _, abJoin := range currentAbilityList {
		if ability.ID == abJoin.AbilityID {
			targetAbility = &abJoin
		}
	}
	if targetAbility == nil {
		return nil, errors.New("ability not found")
	}
	_, err = query.UpdatePlayerAbilityQuantity(ctx, models.UpdatePlayerAbilityQuantityParams{
		Quantity:  int32(quantity),
		PlayerID:  ih.player.ID,
		AbilityID: ability.ID,
	})
	if err != nil {
		return nil, err
	}

	return &ability, nil
}
