package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	dbmigrate "github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// MigrationsHandler implements the database migrations admin page
// (/admin/migrations): see applied/pending state and apply/rollback from the
// panel, using the embedded golang-migrate runner. Destructive actions are
// hard-blocked against the production database.
type MigrationsHandler struct {
	getRunner func() *dbmigrate.Runner // lazy: nil provider result = unavailable
	isProd    bool
}

// NewMigrationsHandler creates a MigrationsHandler. getRunner may return nil
// (the page then renders an "unavailable" state).
func NewMigrationsHandler(getRunner func() *dbmigrate.Runner, isProd bool) *MigrationsHandler {
	return &MigrationsHandler{getRunner: getRunner, isProd: isProd}
}

// runner returns the embedded runner or nil when unavailable.
func (h *MigrationsHandler) runner() *dbmigrate.Runner {
	if h.getRunner == nil {
		return nil
	}
	return h.getRunner()
}

// Page handles GET /admin/migrations.
func (h *MigrationsHandler) Page(c echo.Context) error {
	data := h.loadData()
	return render(c, http.StatusOK, pages.MigrationsPage(data))
}

// Up handles POST /admin/migrations/up — apply all pending migrations.
func (h *MigrationsHandler) Up(c echo.Context) error {
	if h.blocked(c) {
		return nil
	}
	r := h.runner()
	if r == nil {
		return h.migrateError(c, "Migrations runner unavailable — no database DSN configured")
	}
	if err := r.Up(); err != nil {
		return h.migrateError(c, "Migration failed: "+err.Error())
	}
	c.Response().Header().Set("HX-Trigger", toastTrigger("All pending migrations applied", "success"))
	return render(c, http.StatusOK, pages.MigrationsContent(h.loadData()))
}

// Down handles POST /admin/migrations/down — roll back n steps. The user must
// type the name of the migration being rolled back (server-validated), so a
// stray click can't destroy schema.
func (h *MigrationsHandler) Down(c echo.Context) error {
	if h.blocked(c) {
		return nil
	}
	r := h.runner()
	if r == nil {
		return h.migrateError(c, "Migrations runner unavailable — no database DSN configured")
	}

	n := 1
	if v, err := strconv.Atoi(c.FormValue("steps")); err == nil && v > 0 {
		n = v
	}
	if n > 10 {
		n = 10
	}

	last, err := h.lastApplied(r)
	if err != nil {
		return h.migrateError(c, "Failed to determine current migration: "+err.Error())
	}
	if last == nil {
		return h.migrateError(c, "No migrations applied — nothing to roll back")
	}

	if got := c.FormValue("confirm"); got != last.Name {
		c.Response().Header().Set("HX-Trigger", toastTrigger("Type the exact migration name to confirm the rollback", "error"))
		return c.String(http.StatusBadRequest, "confirmation phrase mismatch")
	}

	if err := r.DownSteps(n); err != nil {
		return h.migrateError(c, "Rollback failed: "+err.Error())
	}
	c.Response().Header().Set("HX-Trigger", toastTrigger("Rolled back "+strconv.Itoa(n)+" migration(s)", "success"))
	return render(c, http.StatusOK, pages.MigrationsContent(h.loadData()))
}

// blocked applies the production guard and writes the 403 response itself.
func (h *MigrationsHandler) blocked(c echo.Context) bool {
	if h.isProd {
		c.Response().Header().Set("HX-Trigger", toastTrigger("Production migrations are disabled from the web panel.", "error"))
		_ = c.String(http.StatusForbidden, "migrations blocked against production")
		return true
	}
	return false
}

func (h *MigrationsHandler) migrateError(c echo.Context, msg string) error {
	c.Response().Header().Set("HX-Trigger", toastTrigger(msg, "error"))
	return c.String(http.StatusInternalServerError, msg)
}

// lastApplied returns the highest-version applied migration, or nil.
func (h *MigrationsHandler) lastApplied(r *dbmigrate.Runner) (*dbmigrate.Migration, error) {
	st, err := r.Status()
	if err != nil {
		return nil, err
	}
	for i := len(st) - 1; i >= 0; i-- {
		if st[i].Applied {
			return &st[i], nil
		}
	}
	return nil, nil
}

// loadData builds the page view model.
func (h *MigrationsHandler) loadData() pages.MigrationsPageData {
	data := pages.MigrationsPageData{
		IsProd: h.isProd,
	}
	r := h.runner()
	if r == nil {
		data.Unavailable = true
		return data
	}

	st, err := r.Status()
	if err != nil {
		data.Error = err.Error()
		return data
	}
	for i, m := range st {
		view := pages.MigrationView{
			Version: m.Version,
			Name:    m.Name,
			Applied: m.Applied,
			Dirty:   m.Dirty,
		}
		if i == len(st)-1 {
			view.Last = true
		}
		if !m.Applied {
			data.HasPending = true
		}
		data.Migrations = append(data.Migrations, view)
	}

	if last, err := h.lastApplied(r); err == nil && last != nil {
		data.LastAppliedName = last.Name
	}
	return data
}
