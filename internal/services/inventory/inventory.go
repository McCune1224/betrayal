package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

type PlayerAbilities struct {
	AbilityID   int    `json:"ability_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PlayerItem struct {
	ItemID      int    `json:"item_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PlayerImmunities struct {
	ImmunityID  int    `json:"immunity_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PlayerInventory struct {
	models.Player
	Abilities  PlayerAbilities  `json:"abilities"`
	Items      PlayerItem       `json:"items"`
	Immunities PlayerImmunities `json:"immunities"`
}

type InventoryHandler struct {
	query  *models.Queries
	player models.Player
}

func NewInventoryHandler(playerID int64, db *pgxpool.Pool, adminOnly bool) (*InventoryHandler, error) {
	handler := &InventoryHandler{}
	query := models.New(db)
	player, err := query.GetPlayer(context.Background(), playerID)
	if err != nil {
		return nil, err
	}
	handler.player = player
	return handler, nil
}

// FIXME: The sqlc query here def is not working...
func (ih *InventoryHandler) FetchInventory() (*PlayerInventory, error) {
	log.Println("player: ", ih.player)
	result, err := ih.query.PlayerFoo(context.Background(), ih.player.ID)
	if err != nil {
		return nil, err
	}

	player := &PlayerInventory{
		Player:     ih.player,
		Abilities:  PlayerAbilities{},
		Items:      PlayerItem{},
		Immunities: PlayerImmunities{},
	}

	// Unmarshal the JSON fields
	err = json.Unmarshal(result.AbilityDetails, &player.Abilities)
	if err != nil {
		return nil, errors.New("Error unmarshalling ability details: %v")
	}
	err = json.Unmarshal(result.ItemDetails, &player.Items)
	if err != nil {
		return nil, errors.New("Error unmarshalling item details: %v")
	}
	err = json.Unmarshal(result.ImmunityDetails, &player.Immunities)
	if err != nil {
		return nil, errors.New("Error unmarshalling immunity details: %v")
	}

	return player, nil
}
