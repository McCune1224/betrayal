package api

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
)

// PlayersHandler exposes the player list as explicit API DTOs.
type PlayersHandler struct {
	pool *pgxpool.Pool
}

// NewPlayersHandler creates a player-list API handler backed by the shared pool.
func NewPlayersHandler(pool *pgxpool.Pool) *PlayersHandler {
	return &PlayersHandler{pool: pool}
}

type playerListDTO struct {
	ID        int64  `json:"id"`
	Alive     bool   `json:"alive"`
	Coins     int    `json:"coins"`
	Luck      int    `json:"luck"`
	ItemLimit int    `json:"item_limit"`
	Alignment string `json:"alignment"`
	Role      string `json:"role"`
}

// List returns the current players as stable display-ready API DTOs.
func (h *PlayersHandler) List(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.pool)
	players, err := q.ListPlayer(ctx)
	if err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "players_unavailable", "could not load players", map[string]any{})
		return nil
	}

	result := make([]playerListDTO, len(players))
	for i, player := range players {
		roleName := "Unknown"
		if player.RoleID.Valid {
			if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
				roleName = role.Name
			}
		}
		result[i] = playerListDTO{
			ID:        player.ID,
			Alive:     player.Alive,
			Coins:     int(player.Coins),
			Luck:      int(player.Luck),
			ItemLimit: int(player.ItemLimit),
			Alignment: string(player.Alignment),
			Role:      roleName,
		}
	}

	WriteJSON(c.Response(), http.StatusOK, result)
	return nil
}
