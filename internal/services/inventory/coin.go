package inventory

import (
	"context"
	"strconv"

	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
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

func (ih *InventoryHandler) UpdateCoinBonus(coinBonusStr string) error {
	f64, err := strconv.ParseFloat(coinBonusStr, 64)
	if err != nil {
		return err
	}

	numericCoinBonus, err := util.Numeric(f64)
	if err != nil {
		return err
	}

	query := models.New(ih.pool)
	_, err = query.UpdatePlayerCoinBonus(context.Background(), models.UpdatePlayerCoinBonusParams{
		ID:        ih.player.ID,
		CoinBonus: numericCoinBonus,
	})
	if err != nil {
		return err
	}
	return nil
}
