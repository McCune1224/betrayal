package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/web/templates/partials"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	dbPool         *pgxpool.Pool
	discordSession *discordgo.Session
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(pool *pgxpool.Pool, discord *discordgo.Session) *HealthHandler {
	return &HealthHandler{
		dbPool:         pool,
		discordSession: discord,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Health handles GET /health
func (h *HealthHandler) Health(c echo.Context) error {
	response := HealthResponse{
		Status:   "ok",
		Database: "ok",
	}

	// Check database connectivity
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	if err := h.dbPool.Ping(ctx); err != nil {
		response.Status = "degraded"
		response.Database = "error"
		return c.JSON(http.StatusServiceUnavailable, response)
	}

	return c.JSON(http.StatusOK, response)
}

// HealthStatusPartial handles GET /health/status (HTMX partial)
func (h *HealthHandler) HealthStatusPartial(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Check database status
	dbStatus := "ok"
	if err := h.dbPool.Ping(ctx); err != nil {
		dbStatus = "offline"
	}

	// Check Discord bot status
	botStatus := "offline"
	if h.discordSession != nil && h.discordSession.State != nil && h.discordSession.State.User != nil {
		botStatus = "online"
	}

	// Get current timestamp
	lastChecked := time.Now().Format("15:04:05 MST")

	component := partials.HealthStatus(botStatus, dbStatus, lastChecked)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
