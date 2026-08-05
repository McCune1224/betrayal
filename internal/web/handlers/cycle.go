package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// CycleHandler handles the game cycle admin page (/cycle)
type CycleHandler struct {
	dbPool *pgxpool.Pool
}

// NewCycleHandler creates a new CycleHandler
func NewCycleHandler(pool *pgxpool.Pool) *CycleHandler {
	return &CycleHandler{dbPool: pool}
}

// Page handles GET /cycle — current cycle + advance/set controls
func (h *CycleHandler) Page(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	data, err := h.loadCycleData(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load game cycle")
	}

	return render(c, http.StatusOK, pages.Cycle(data))
}

// Advance handles POST /cycle/advance — move to the next phase.
// Mirrors the Discord /cycle next command's phase transitions (Day n ->
// Elimination n -> Day n+1), but DB-only: the web panel does not broadcast
// cycle messages to Discord channels.
func (h *CycleHandler) Advance(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)
	curr, err := q.GetCycle(ctx)
	if err != nil {
		return h.cycleError(c, "Failed to load current cycle")
	}

	var updated models.GameCycle
	switch {
	case curr.Day == 0:
		// Day 0 is game start: Day 0 -> Day 1
		updated, err = q.UpdateCycle(ctx, models.UpdateCycleParams{
			IsElimination: false,
			Day:           1,
			ID:            curr.ID,
		})
	case curr.IsElimination:
		// Elimination n -> Day n+1
		updated, err = q.UpdateCycle(ctx, models.UpdateCycleParams{
			IsElimination: false,
			Day:           curr.Day + 1,
			ID:            curr.ID,
		})
	default:
		// Day n -> Elimination n
		updated, err = q.UpdateCycle(ctx, models.UpdateCycleParams{
			IsElimination: true,
			Day:           curr.Day,
			ID:            curr.ID,
		})
	}
	if err != nil {
		return h.cycleError(c, "Failed to advance cycle")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Cycle advanced to `+formatCyclePhase(updated)+`", "type": "success"}}`)
	return render(c, http.StatusOK, pages.CycleStatusCard(toCycleData(updated)))
}

// Set handles POST /cycle/set — hard set the phase and day number.
// Mirrors the Discord /cycle set command (no Discord broadcast).
func (h *CycleHandler) Set(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	phase := c.FormValue("phase")
	numberStr := c.FormValue("number")

	day, err := strconv.ParseInt(numberStr, 10, 32)
	if err != nil || day < 0 {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid day number", "type": "error"}}`)
		return c.String(http.StatusBadRequest, "Invalid day number")
	}

	isElimination := phase == "Elimination"
	if phase != "Day" && phase != "Elimination" {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid phase (must be Day or Elimination)", "type": "error"}}`)
		return c.String(http.StatusBadRequest, "Invalid phase")
	}

	q := models.New(h.dbPool)
	curr, err := q.GetCycle(ctx)
	if err != nil {
		return h.cycleError(c, "Failed to load current cycle")
	}

	updated, err := q.UpdateCycle(ctx, models.UpdateCycleParams{
		IsElimination: isElimination,
		Day:           int32(day),
		ID:            curr.ID,
	})
	if err != nil {
		return h.cycleError(c, "Failed to set cycle")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Cycle set to `+formatCyclePhase(updated)+`", "type": "success"}}`)
	return render(c, http.StatusOK, pages.CycleStatusCard(toCycleData(updated)))
}

// loadCycleData fetches the current cycle for the full page render
func (h *CycleHandler) loadCycleData(ctx context.Context) (pages.CycleData, error) {
	q := models.New(h.dbPool)
	cycle, err := q.GetCycle(ctx)
	if err != nil {
		return pages.CycleData{}, err
	}
	return toCycleData(cycle), nil
}

func toCycleData(cycle models.GameCycle) pages.CycleData {
	phase := "Day"
	if cycle.IsElimination {
		phase = "Elimination"
	}
	return pages.CycleData{
		Phase: phase,
		Day:   int(cycle.Day),
	}
}

func formatCyclePhase(cycle models.GameCycle) string {
	return toCycleData(cycle).String()
}

// cycleError sets an HTMX error toast and returns a plain-text 500
func (h *CycleHandler) cycleError(c echo.Context, msg string) error {
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "`+msg+`", "type": "error"}}`)
	return c.String(http.StatusInternalServerError, msg)
}
