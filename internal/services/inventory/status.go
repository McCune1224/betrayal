package inventory

import (
	"context"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddStatus(statusName string, quantity int32) (*models.Status, error) {
	query := models.New(ih.pool)
	status, err := query.GetStatusByFuzzy(context.Background(), statusName)
	if err != nil {
		return nil, err
	}
	err = query.UpsertPlayerStatusJoin(context.TODO(), models.UpsertPlayerStatusJoinParams{
		PlayerID: ih.player.ID,
		StatusID: status.ID,
		Quantity: quantity,
	})
	if err != nil {
		return nil, err
	}

	return &status, nil

}

func (ih *InventoryHandler) RemoveStatus(statusName string, quantity int32) (*models.Status, error) {
	query := models.New(ih.pool)
	status, err := query.GetStatusByFuzzy(context.Background(), statusName)
	if err != nil {
		return nil, err
	}
	items, err := query.ListPlayerItemInventory(context.Background(), ih.player.ID)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.ID == status.ID {
			if i.Quantity-quantity <= 0 {
				err = query.DeletePlayerStatus(context.Background(), models.DeletePlayerStatusParams{
					PlayerID: ih.player.ID,
					StatusID: status.ID,
				})
			} else {
				_, err = query.UpdatePlayerStatusQuantity(context.Background(), models.UpdatePlayerStatusQuantityParams{
					PlayerID: ih.player.ID,
					StatusID: status.ID,
					Quantity: i.Quantity - 1,
				})
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &status, err
}
