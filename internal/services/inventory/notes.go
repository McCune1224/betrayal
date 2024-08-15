package inventory

import (
	"context"
	"errors"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) CreatePlayerNote(playerID int64, info string) (*models.PlayerNote, error) {
	q := models.New(ih.pool)
	dbCtx := context.Background()

	nextPosition, _ := q.GetPlayerNoteCount(dbCtx, playerID)
	note, err := q.CreatePlayerNote(dbCtx, models.CreatePlayerNoteParams{
		PlayerID: playerID,
		Position: int32(nextPosition) + 1,
		Info:     info,
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (ih *InventoryHandler) UpdatePlayerNote(playerID int64, position int, info string) (*models.PlayerNote, error) {
	q := models.New(ih.pool)
	dbCtx := context.Background()

	totalPositions, _ := q.GetPlayerNoteCount(dbCtx, playerID)
	if position > int(totalPositions) || position < 1 {
		return nil, errors.New("position is greater than total positions")
	}

	note, err := q.UpdatePlayerNoteByPosition(dbCtx, models.UpdatePlayerNoteByPositionParams{
		PlayerID: playerID,
		Position: int32(totalPositions),
		Info:     info,
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (ih *InventoryHandler) DeletePlayerNote(playerID int64, position int) error {
	q := models.New(ih.pool)
	dbCtx := context.Background()

	nextPosition, _ := q.GetPlayerNoteCount(dbCtx, playerID)
	totalPositions, _ := q.GetPlayerNoteCount(dbCtx, playerID)
	if position > int(totalPositions) || position < 1 {
		return errors.New("position is greater than total positions")
	}

	err := q.DeletePlayerNoteByPosition(dbCtx, models.DeletePlayerNoteByPositionParams{
		PlayerID: playerID,
		Position: int32(nextPosition),
	})
	if err != nil {
		return err
	}
	return nil
}

func (ih *InventoryHandler) GetPlayerNote(playerID int64, position int) (*models.PlayerNote, error) {
	q := models.New(ih.pool)
	dbCtx := context.Background()

	note, err := q.GetPlayerNoteByPosition(dbCtx, models.GetPlayerNoteByPositionParams{
		PlayerID: playerID,
		Position: int32(position),
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}
