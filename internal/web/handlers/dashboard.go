package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	dbPool *pgxpool.Pool
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{dbPool: pool}
}

// Dashboard handles GET /
func (h *DashboardHandler) Dashboard(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	// Get current cycle
	cycle, err := q.GetCycle(ctx)
	cyclePhase := "Day"
	cycleNumber := 1
	if err == nil {
		if cycle.IsElimination {
			cyclePhase = "Elimination"
		} else {
			cyclePhase = "Day"
		}
		cycleNumber = int(cycle.Day)
	}

	// Get player counts
	players, err := q.ListPlayer(ctx)
	if err != nil {
		players = []models.Player{}
	}

	var alive, dead int
	for _, p := range players {
		if p.Alive {
			alive++
		} else {
			dead++
		}
	}

	data := pages.DashboardData{
		CyclePhase:   cyclePhase,
		CycleNumber:  cycleNumber,
		PlayersAlive: alive,
		PlayersDead:  dead,
		TotalPlayers: len(players),
	}

	return render(c, http.StatusOK, pages.Dashboard(data))
}

// formatPlayerID formats a Discord user ID for display
func formatPlayerID(id int64) string {
	return fmt.Sprintf("%d", id)
}
