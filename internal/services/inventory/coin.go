package inventory

import (
	"context"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) AddCoin(quantity int32) error {
	query := models.New(ih.pool)
	_, err := query.UpdatePlayerCoins(context.Background(), models.UpdatePlayerCoinsParams{
		ID:    ih.player.ID,
		Coins: int32(quantity) + ih.player.Coins,
	})
	return err
}

func (ih *InventoryHandler) RemoveCoin(quantity int32) error {
	query := models.New(ih.pool)
	diff := ih.player.Coins - int32(quantity)
	if diff < 0 {
		diff = 0
	}
	_, err := query.UpdatePlayerCoins(context.Background(), models.UpdatePlayerCoinsParams{
		ID:    ih.player.ID,
		Coins: diff,
	})
	return err
}

func (ih *InventoryHandler) SetCoin(quantity int32) error {
	if quantity < 0 {
		quantity = 0
	}
	query := models.New(ih.pool)
	_, err := query.UpdatePlayerCoins(context.Background(), models.UpdatePlayerCoinsParams{
		ID:    ih.player.ID,
		Coins: quantity,
	})
	return err
}
