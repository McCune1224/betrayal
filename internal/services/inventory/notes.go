package inventory

import (
	"errors"

	"github.com/mccune1224/betrayal/internal/models"
)

func (ih *InventoryHandler) CreatePlayerNote(playerID int64, info string) (*models.PlayerNote, error) {
	q := models.New(ih.pool)
	ctx, cancel := dbCtx()
	defer cancel()

	notes, err := q.ListPlayerNote(ctx, playerID)
	if err != nil {
		return nil, err
	}
	nextPosition := 0
	for _, note := range notes {
		if int(note.Position) > nextPosition {
			nextPosition = int(note.Position)
		}
	}
	note, err := q.CreatePlayerNote(ctx, models.CreatePlayerNoteParams{
		PlayerID: playerID,
		Position: int32(nextPosition + 1),
		Info:     info,
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (ih *InventoryHandler) UpdatePlayerNote(playerID int64, position int, info string) (*models.PlayerNote, error) {
	q := models.New(ih.pool)
	ctx, cancel := dbCtx()
	defer cancel()

	totalPositions, err := q.GetPlayerNoteCount(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if position > int(totalPositions) || position < 1 {
		return nil, errors.New("position is greater than total positions")
	}

	note, err := q.UpdatePlayerNoteByPosition(ctx, models.UpdatePlayerNoteByPositionParams{
		PlayerID: playerID,
		Position: int32(position),
		Info:     info,
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (ih *InventoryHandler) DeletePlayerNote(playerID int64, position int) error {
	q := models.New(ih.pool)
	ctx, cancel := dbCtx()
	defer cancel()

	totalPositions, err := q.GetPlayerNoteCount(ctx, playerID)
	if err != nil {
		return err
	}
	if position > int(totalPositions) || position < 1 {
		return errors.New("position is greater than total positions")
	}

	err = q.DeletePlayerNoteByPosition(ctx, models.DeletePlayerNoteByPositionParams{
		PlayerID: playerID,
		Position: int32(position),
	})
	if err != nil {
		return err
	}
	return nil
}

func (ih *InventoryHandler) GetPlayerNote(playerID int64, position int) (*models.PlayerNote, error) {
	q := models.New(ih.pool)
	ctx, cancel := dbCtx()
	defer cancel()

	note, err := q.GetPlayerNoteByPosition(ctx, models.GetPlayerNoteByPositionParams{
		PlayerID: playerID,
		Position: int32(position),
	})
	if err != nil {
		return nil, err
	}
	return &note, nil
}
