package inventory

import (
	"context"
	"errors"
	"log"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddAbility(abilityName string, quantity int32) (*models.AbilityInfo, error) {
	query := models.New(ih.pool)
	ability, err := query.GetAbilityInfoByFuzzy(context.Background(), abilityName)
	if err != nil {
		return nil, err
	}
	currentAbilityIds, _ := query.ListPlayerAbilityJoin(context.Background(), ih.player.ID)

	for _, abilityId := range currentAbilityIds {
		if ability.ID == abilityId.AbilityID {
			return nil, errors.New("ability already added")
		}
	}

	if quantity == 0 {
		quantity = ability.DefaultCharges
	}

	_, err = query.CreatePlayerAbilityJoin(context.Background(), models.CreatePlayerAbilityJoinParams{
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
	query := models.New(ih.pool)
	ability, err := query.GetAbilityInfoByFuzzy(context.Background(), abilityName)
	if err != nil {
		return nil, err
	}
	err = query.DeletePlayerAbility(context.Background(), models.DeletePlayerAbilityParams{
		PlayerID:  ih.player.ID,
		AbilityID: ability.ID,
	})
	return &ability, err
}

func (ih *InventoryHandler) UpdateAbility(abilityName string, quantity int) (*models.AbilityInfo, error) {
	query := models.New(ih.pool)
	ability, err := query.GetAbilityInfoByFuzzy(context.Background(), abilityName)
	if err != nil {
		return nil, err
	}

	currentAbilityList, _ := query.ListPlayerAbilityJoin(context.Background(), ih.player.ID)
	log.Println(currentAbilityList)
	targetAbility := &models.PlayerAbility{}
	for _, abJoin := range currentAbilityList {
		if ability.ID == abJoin.AbilityID {
			targetAbility = &abJoin
		}
	}
	if targetAbility == nil {
		return nil, errors.New("ability not found")
	}
	if quantity < 0 {
		quantity = 0
	}
	_, err = query.UpdatePlayerAbilityQuantity(context.Background(), models.UpdatePlayerAbilityQuantityParams{
		Quantity:  int32(quantity),
		PlayerID:  ih.player.ID,
		AbilityID: ability.ID,
	})
	if err != nil {
		return nil, err
	}

	return &ability, nil
}
