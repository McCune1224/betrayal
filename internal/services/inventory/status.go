package inventory

import (
	"errors"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddStatus(statusName string, quantity int32) (*models.Status, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	ctx, cancel := dbCtx()
	defer cancel()

	query := models.New(ih.pool)
	status, err := query.GetStatusByFuzzy(ctx, statusName)
	if err != nil {
		return nil, err
	}
	err = query.UpsertPlayerStatusJoin(ctx, models.UpsertPlayerStatusJoinParams{
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
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	ctx, cancel := dbCtx()
	defer cancel()

	query := models.New(ih.pool)
	status, err := query.GetStatusByFuzzy(ctx, statusName)
	if err != nil {
		return nil, err
	}
	statuses, err := query.ListPlayerStatusInventory(ctx, ih.player.ID)
	if err != nil {
		return nil, err
	}
	for _, i := range statuses {
		if i.ID == status.ID {
			if i.Quantity-quantity <= 0 {
				err = query.DeletePlayerStatus(ctx, models.DeletePlayerStatusParams{
					PlayerID: ih.player.ID,
					StatusID: status.ID,
				})
			} else {
				_, err = query.UpdatePlayerStatusQuantity(ctx, models.UpdatePlayerStatusQuantityParams{
					PlayerID: ih.player.ID,
					StatusID: status.ID,
					Quantity: i.Quantity - quantity,
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
