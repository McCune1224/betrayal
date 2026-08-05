package inventory

import (
	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddItem(itemName string, quantity int32) (*models.Item, error) {
	ctx, cancel := dbCtx()
	defer cancel()

	query := models.New(ih.pool)
	item, err := query.GetItemByFuzzy(ctx, itemName)
	if err != nil {
		return nil, err
	}
	err = query.UpsertPlayerItemJoin(ctx, models.UpsertPlayerItemJoinParams{
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
	ctx, cancel := dbCtx()
	defer cancel()

	query := models.New(ih.pool)
	item, err := query.GetItemByFuzzy(ctx, itemName)
	if err != nil {
		return nil, err
	}
	items, err := query.ListPlayerItemInventory(ctx, ih.player.ID)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.ID == item.ID {
			if i.Quantity-quantity <= 0 {
				err = query.DeletePlayerItem(ctx, models.DeletePlayerItemParams{
					PlayerID: ih.player.ID,
					ItemID:   item.ID,
				})
			} else {
				_, err = query.UpdatePlayerItemQuantity(ctx, models.UpdatePlayerItemQuantityParams{
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
