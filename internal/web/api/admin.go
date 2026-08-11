package api

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	dbmigrate "github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/gamereset"
	"github.com/mccune1224/betrayal/internal/web/railway"
)

type AuditEntryDTO struct {
	ID         int64     `json:"id"`
	Command    string    `json:"command"`
	Arguments  string    `json:"arguments,omitempty"`
	Username   string    `json:"username,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	ChannelID  string    `json:"channel_id,omitempty"`
	ExecutedAt time.Time `json:"executed_at"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
}
type AuditDTO struct {
	Entries []AuditEntryDTO `json:"entries"`
}
type AdminHandler struct {
	pool    *pgxpool.Pool
	railway *railway.Client
	runner  func() *dbmigrate.Runner
	reset   *gamereset.Service
}

func NewAdminHandler(pool *pgxpool.Pool, railwayClient *railway.Client, runner func() *dbmigrate.Runner, reset *gamereset.Service) *AdminHandler {
	return &AdminHandler{pool: pool, railway: railwayClient, runner: runner, reset: reset}
}
func (h *AdminHandler) Audit(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	logs, err := models.New(h.pool).ListRecentCommands(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "audit_unavailable", "could not load audit log", nil)
		return nil
	}
	out := make([]AuditEntryDTO, 0, len(logs))
	for _, l := range logs {
		out = append(out, AuditEntryDTO{ID: l.ID, Command: l.CommandName, Arguments: formatArgs(l.CommandArguments), Username: l.Username, UserID: l.UserID, ChannelID: l.ChannelID.String, ExecutedAt: l.Timestamp.Time, Status: l.Status.String, Error: l.ErrorMessage.String})
	}
	WriteJSON(c.Response(), 200, AuditDTO{Entries: out})
	return nil
}
func (h *AdminHandler) Migrations(c echo.Context) error {
	r := h.runner()
	if r == nil {
		WriteError(c.Response(), 503, "migrations_unavailable", "migrations runner unavailable", nil)
		return nil
	}
	status, err := r.Status()
	if err != nil {
		WriteError(c.Response(), 500, "migrations_unavailable", "could not load migration status", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, map[string]any{"migrations": status})
	return nil
}
func (h *AdminHandler) MigrationUp(c echo.Context) error {
	r := h.runner()
	if r == nil {
		WriteError(c.Response(), 503, "migrations_unavailable", "migrations runner unavailable", nil)
		return nil
	}
	if err := r.Up(); err != nil {
		WriteError(c.Response(), 500, "migration_failed", "could not apply migrations", nil)
		return nil
	}
	return h.Migrations(c)
}

type MigrationDownRequest struct {
	Steps   int    `json:"steps"`
	Confirm string `json:"confirm"`
}

func (h *AdminHandler) MigrationDown(c echo.Context) error {
	r := h.runner()
	if r == nil {
		WriteError(c.Response(), 503, "migrations_unavailable", "migrations runner unavailable", nil)
		return nil
	}
	var req MigrationDownRequest
	if err := c.Bind(&req); err != nil {
		WriteError(c.Response(), 400, "invalid_request", "invalid JSON body", nil)
		return nil
	}
	if req.Steps <= 0 {
		req.Steps = 1
	}
	if req.Steps > 10 {
		WriteError(c.Response(), 400, "invalid_request", "steps must be between 1 and 10", nil)
		return nil
	}
	status, err := r.Status()
	if err != nil {
		WriteError(c.Response(), 500, "migrations_unavailable", "could not determine current migration", nil)
		return nil
	}
	last := ""
	for i := len(status) - 1; i >= 0; i-- {
		if status[i].Applied {
			last = status[i].Name
			break
		}
	}
	if last == "" {
		WriteError(c.Response(), 400, "migration_unavailable", "no migrations are applied", nil)
		return nil
	}
	if req.Confirm != last {
		WriteError(c.Response(), 400, "confirmation_required", "confirm the latest migration name", map[string]any{"latest_migration": last})
		return nil
	}
	if err := r.DownSteps(req.Steps); err != nil {
		WriteError(c.Response(), 500, "migration_failed", "could not roll back migrations", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, map[string]any{"status": "rolled_back", "steps": req.Steps, "migration": last})
	return nil
}

type ResetRequest struct {
	Confirm    string `json:"confirm"`
	Understand bool   `json:"understand"`
}

func (h *AdminHandler) ResetPreview(c echo.Context) error {
	summary, err := h.reset.Preview(c.Request().Context())
	if err != nil {
		WriteError(c.Response(), 500, "reset_unavailable", "could not preview reset", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, map[string]any{"confirmation": "RESET BETRAYAL GAME", "summary": summary})
	return nil
}
func (h *AdminHandler) ResetExecute(c echo.Context) error {
	var req ResetRequest
	if err := c.Bind(&req); err != nil {
		WriteError(c.Response(), 400, "invalid_request", "invalid JSON body", nil)
		return nil
	}
	if req.Confirm != "RESET BETRAYAL GAME" || !req.Understand {
		WriteError(c.Response(), 400, "confirmation_required", "type RESET BETRAYAL GAME and acknowledge the reset", nil)
		return nil
	}
	result, err := h.reset.Execute(c.Request().Context())
	if err != nil {
		WriteError(c.Response(), 500, "reset_failed", "reset cancelled; no database changes were committed", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, result)
	return nil
}
func (h *AdminHandler) Redeploy(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()
	id, err := h.railway.GetLatestDeployment(ctx)
	if err != nil {
		WriteError(c.Response(), 502, "redeploy_failed", "could not find latest deployment", nil)
		return nil
	}
	if err := h.railway.RestartDeployment(ctx, id); err != nil {
		WriteError(c.Response(), 502, "redeploy_failed", "could not restart deployment", nil)
		return nil
	}
	WriteJSON(c.Response(), 202, map[string]string{"status": "started"})
	return nil
}
func formatArgs(v []byte) string {
	if len(v) == 0 {
		return ""
	}
	return string(v)
}
