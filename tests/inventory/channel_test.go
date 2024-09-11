package inventory

import (
	"context"
	"log"
	"time"

	"github.com/mccune1224/betrayal/internal/models"
)

// 206268866714796032 -- Alex

func (its *InventoryTestSuite) TestChannel() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	query := models.New(its.DB)
	player, err := query.GetPlayer(ctx, 206268866714796032)
	if err != nil {
		its.Fail(err.Error())
	}

	abilityChan := make(chan []models.ListPlayerAbilityInventoryRow, 1)
	itemCh := make(chan []models.ListPlayerItemInventoryRow, 1)
	statusChan := make(chan []models.ListPlayerStatusRow, 1)
	immunityChan := make(chan []models.ListPlayerImmunityRow, 1)

	now := time.Now()
	go dbTask(ctx, abilityChan, func() ([]models.ListPlayerAbilityInventoryRow, error) {
		return query.ListPlayerAbilityInventory(ctx, player.ID)
	})

	go dbTask(ctx, itemCh, func() ([]models.ListPlayerItemInventoryRow, error) {
		return query.ListPlayerItemInventory(ctx, player.ID)
	})

	go dbTask(ctx, statusChan, func() ([]models.ListPlayerStatusRow, error) {
		return query.ListPlayerStatus(ctx, player.ID)
	})

	go dbTask(ctx, immunityChan, func() ([]models.ListPlayerImmunityRow, error) {
		return query.ListPlayerImmunity(ctx, player.ID)
	})

	abResult := <-abilityChan
	itemResult := <-itemCh
	immunityResult := <-immunityChan
	statusResult := <-statusChan
	log.Println("HIT", len(abResult), len(itemResult), len(immunityResult), len(statusResult), time.Since(now))
}

func dbTask[T any](ctx context.Context, resultChan chan T, dbFunc func() (T, error)) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	select {
	case <-ctx.Done():
		close(resultChan)
		cancel()
	default:
		result, err := dbFunc()
		if err != nil {
			close(resultChan)
			cancel()
			return
		}
		resultChan <- result
		return
	}
}
