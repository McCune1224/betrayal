package inventory

import (
	"context"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddItem(itemName string, quantity int32) (*models.Item, error) {
	query := models.New(ih.pool)
	item, err := query.GetItemByFuzzy(context.Background(), itemName)
	if err != nil {
		return nil, err
	}
	err = query.UpsertPlayerItemJoin(context.TODO(), models.UpsertPlayerItemJoinParams{
		PlayerID: ih.player.ID,
		ItemID:   item.ID,
		Quantity: quantity,
	})
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (ih *InventoryHandler) RemoveItem(itemName string, quantity int32) (*models.Item, error) {
	query := models.New(ih.pool)
	item, err := query.GetItemByFuzzy(context.Background(), itemName)
	if err != nil {
		return nil, err
	}
	items, err := query.ListPlayerItemInventory(context.Background(), ih.player.ID)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.ID == item.ID {
			if i.Quantity-quantity <= 0 {
				err = query.DeletePlayerItem(context.Background(), models.DeletePlayerItemParams{
					PlayerID: ih.player.ID,
					ItemID:   item.ID,
				})
			} else {
				_, err = query.UpdatePlayerItemQuantity(context.Background(), models.UpdatePlayerItemQuantityParams{
					PlayerID: ih.player.ID,
					ItemID:   item.ID,
					Quantity: i.Quantity - 1,
				})
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &item, err
}
