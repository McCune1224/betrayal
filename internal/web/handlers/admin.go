package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/railway"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	dbPool        *pgxpool.Pool
	railwayClient *railway.Client
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(pool *pgxpool.Pool, railwayClient *railway.Client) *AdminHandler {
	return &AdminHandler{
		dbPool:        pool,
		railwayClient: railwayClient,
	}
}

// Redeploy handles POST /admin/redeploy
func (h *AdminHandler) Redeploy(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	// Get latest deployment
	deploymentID, err := h.railwayClient.GetLatestDeployment(ctx)
	if err != nil {
		// Return HTMX-friendly error response
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to get deployment: `+err.Error()+`", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to get deployment")
	}

	// Restart deployment
	if err := h.railwayClient.RestartDeployment(ctx, deploymentID); err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to restart: `+err.Error()+`", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to restart deployment")
	}

	// Return success with HTMX trigger for toast notification
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Redeploy triggered successfully!", "type": "success"}}`)
	return c.String(http.StatusOK, "Redeploy triggered")
}

// AuditLogs handles GET /admin/audit
func (h *AdminHandler) AuditLogs(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	// Get recent audit logs (last 100 from recent hour)
	logs, err := q.ListRecentCommands(ctx)
	if err != nil {
		logs = []models.CommandAudit{}
	}

	// Convert to view model
	auditData := make([]pages.AuditLogEntry, len(logs))
	for i, log := range logs {
		auditData[i] = pages.AuditLogEntry{
			ID:         int(log.ID),
			Command:    log.CommandName,
			Username:   log.Username,
			UserID:     log.UserID,
			ChannelID:  log.ChannelID.String,
			ExecutedAt: log.Timestamp.Time,
			Status:     log.Status.String,
			ErrorMsg:   log.ErrorMessage.String,
		}
	}

	return render(c, http.StatusOK, pages.AuditLogs(auditData))
}
