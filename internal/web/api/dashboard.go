package api

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
)

// DashboardHandler exposes the game-state summary used by the admin dashboard.
type DashboardHandler struct {
	pool *pgxpool.Pool
}

// NewDashboardHandler creates a dashboard API handler backed by the shared pool.
func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{pool: pool}
}

// Dashboard returns the current cycle and player-life totals as an explicit API DTO.
func (h *DashboardHandler) Dashboard(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.pool)
	cycle, err := q.GetCycle(ctx)
	if err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "dashboard_unavailable", "could not load dashboard", map[string]any{})
		return nil
	}
	players, err := q.ListPlayer(ctx)
	if err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "dashboard_unavailable", "could not load dashboard", map[string]any{})
		return nil
	}

	phase := "Day"
	if cycle.IsElimination {
		phase = "Elimination"
	}
	alive := 0
	for _, player := range players {
		if player.Alive {
			alive++
		}
	}

	WriteJSON(c.Response(), http.StatusOK, struct {
		Cycle struct {
			Phase  string `json:"phase"`
			Number int    `json:"number"`
		} `json:"cycle"`
		Players struct {
			Alive int `json:"alive"`
			Dead  int `json:"dead"`
			Total int `json:"total"`
		} `json:"players"`
	}{
		Cycle: struct {
			Phase  string `json:"phase"`
			Number int    `json:"number"`
		}{Phase: phase, Number: int(cycle.Day)},
		Players: struct {
			Alive int `json:"alive"`
			Dead  int `json:"dead"`
			Total int `json:"total"`
		}{Alive: alive, Dead: len(players) - alive, Total: len(players)},
	})
	return nil
}
