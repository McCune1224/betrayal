package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// SyncHandler implements the spreadsheet sync page (/sync): fetch the four
// Google Sheets CSVs, preview a create/update/skip diff, and apply it
// transactionally. Mirrors the archived cmd/data-entry CLI, but with a
// validate-before-apply flow and a sync_run audit trail.
type SyncHandler struct {
	dbPool          *pgxpool.Pool
	svc             *datasync.Service
	isProd          bool
	allowMutations  bool // WEB_ALLOW_PROD_MUTATIONS=true overrides the prod block
}

// NewSyncHandler creates a SyncHandler. isProd/allowMutations implement the
// production guard: apply is hard-blocked on the prod pooler unless explicitly
// allowed.
func NewSyncHandler(pool *pgxpool.Pool, svc *datasync.Service, isProd, allowMutations bool) *SyncHandler {
	return &SyncHandler{dbPool: pool, svc: svc, isProd: isProd, allowMutations: allowMutations}
}

// Page handles GET /sync — sources + run history (no diff until preview).
func (h *SyncHandler) Page(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()

	data, err := h.loadPageData(ctx)
	if err != nil {
		return h.syncError(c, "Failed to load sync page: "+err.Error())
	}
	data.IsProd = h.isProd
	data.AllowMutations = h.allowMutations || !h.isProd
	return render(c, http.StatusOK, pages.SyncPage(data))
}

// Preview handles POST /sync/preview — fetch every enabled source, diff
// against the database, and render the changes for validation. Read-only
// (never blocked by the prod guard).
func (h *SyncHandler) Preview(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 90*time.Second)
	defer cancel()

	data, err := h.loadPageData(ctx)
	if err != nil {
		return h.syncError(c, "Failed to load sync page: "+err.Error())
	}

	for _, src := range data.Sources {
		if !src.Enabled || src.Url == "" {
			continue
		}
		diff, err := h.buildDiff(ctx, src)
		if err != nil {
			data.Diffs = append(data.Diffs, pages.SourceDiff{
				SourceID:   src.ID,
				SourceName: src.Name,
				Kind:       src.Kind,
				Error:      err.Error(),
			})
			_ = h.svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusFailed, "web", err.Error(), nil)
			continue
		}
		data.Diffs = append(data.Diffs, diff)
		_ = h.svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusPreview, "web", "", actionCounts(diff.Counts))
	}

	data.IsProd = h.isProd
	data.AllowMutations = h.allowMutations || !h.isProd
	return render(c, http.StatusOK, pages.SyncContent(data))
}

// Apply handles POST /sync/apply — re-fetch + re-plan the given source and
// apply it in one transaction (stateless: the plan is re-derived at apply
// time, so the latest sheet state wins). Prod-guarded.
func (h *SyncHandler) Apply(c echo.Context) error {
	if h.isProd && !h.allowMutations {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Blocked: connected to the PRODUCTION database. Set WEB_ALLOW_PROD_MUTATIONS=true to enable sync applies here.", "type": "error"}}`)
		return c.String(http.StatusForbidden, "sync apply blocked against production")
	}

	sourceID, err := strconv.ParseInt(c.FormValue("source_id"), 10, 32)
	if err != nil {
		return h.badRequest(c, "Invalid source")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 90*time.Second)
	defer cancel()

	src, err := h.svc.ListSources(ctx)
	if err != nil {
		return h.syncError(c, "Failed to load sources")
	}
	var target models.SyncSource
	for _, s := range src {
		if s.ID == int32(sourceID) {
			target = s
		}
	}
	if target.ID == 0 {
		return h.badRequest(c, "Unknown source")
	}

	diff, err := h.buildDiff(ctx, pages.SyncSourceView{ID: target.ID, Name: target.Name, Kind: target.Kind, Alignment: target.Alignment, Url: target.Url, Enabled: target.Enabled})
	if err != nil {
		_ = h.svc.RecordRun(ctx, &target.ID, target.Name, datasync.RunStatusFailed, "web", err.Error(), nil)
		return h.syncError(c, "Preview failed before apply: "+err.Error())
	}

	if err := h.applyDiff(ctx, target); err != nil {
		_ = h.svc.RecordRun(ctx, &target.ID, target.Name, datasync.RunStatusFailed, "web", err.Error(), nil)
		return h.syncError(c, "Apply failed: "+err.Error())
	}

	_ = h.svc.RecordRun(ctx, &target.ID, target.Name, datasync.RunStatusApplied, "web", "", actionCounts(diff.Counts))

	data, err := h.loadPageData(ctx)
	if err != nil {
		return h.syncError(c, "Failed to reload sync page: "+err.Error())
	}
	data.IsProd = h.isProd
	data.AllowMutations = h.allowMutations || !h.isProd
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Applied "+`+strconv.Quote(target.Name)+` — see run history", "type": "success"}}`)
	return render(c, http.StatusOK, pages.SyncContent(data))
}

// UpdateSource handles POST /sync/sources/:id — edit a source's URL or
// enabled flag.
func (h *SyncHandler) UpdateSource(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return h.badRequest(c, "Invalid source id")
	}

	url := strings.TrimSpace(c.FormValue("url"))
	enabled := c.FormValue("enabled") == "on" || c.FormValue("enabled") == "true"
	if enabled && url == "" {
		return h.badRequest(c, "URL is required for an enabled source")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	if _, err := h.svc.UpdateSource(ctx, int32(id), url, enabled); err != nil {
		return h.syncError(c, "Failed to update source: "+err.Error())
	}

	data, err := h.loadPageData(ctx)
	if err != nil {
		return h.syncError(c, "Failed to reload sync page: "+err.Error())
	}
	data.IsProd = h.isProd
	data.AllowMutations = h.allowMutations || !h.isProd
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Source updated", "type": "success"}}`)
	return render(c, http.StatusOK, pages.SyncContent(data))
}

// buildDiff fetches + parses + plans one source into a renderable diff.
func (h *SyncHandler) buildDiff(ctx context.Context, src pages.SyncSourceView) (pages.SourceDiff, error) {
	diff := pages.SourceDiff{
		SourceID:   src.ID,
		SourceName: src.Name,
		Kind:       src.Kind,
		Counts:     map[string]int{},
	}

	body, err := h.svc.Fetch(ctx, models.SyncSource{ID: src.ID, Name: src.Name, Kind: src.Kind, Alignment: src.Alignment, Url: src.Url, Enabled: src.Enabled})
	if err != nil {
		return diff, err
	}
	defer body.Close()

	q := models.New(h.dbPool)

	switch src.Kind {
	case "roles":
		docs, warnings, err := datasync.ParseRolesCSV(body, models.Alignment(src.Alignment))
		if err != nil {
			return diff, err
		}
		plan, err := datasync.PlanRoles(ctx, q, models.Alignment(src.Alignment), docs)
		if err != nil {
			return diff, err
		}
		plan.Warnings = append(plan.Warnings, warnings...)
		for _, rp := range plan.Roles {
			diff.Entries = append(diff.Entries, pages.DiffEntry{
				Kind: "role", Name: rp.Doc.Name, Action: string(rp.Action), Changes: rp.Changes, Nested: false,
			})
			for _, ap := range rp.Abilities {
				diff.Entries = append(diff.Entries, pages.DiffEntry{
					Kind: "ability", Name: ap.Doc.Name, Action: string(ap.Action), Changes: ap.Changes, Nested: true,
				})
			}
			for _, pp := range rp.Perks {
				diff.Entries = append(diff.Entries, pages.DiffEntry{
					Kind: "passive", Name: pp.Doc.Name, Action: string(pp.Action), Changes: pp.Changes, Nested: true,
				})
			}
		}
		diff.Warnings = plan.Warnings
		diff.Counts = countsToStrings(plan.Counts)

	case "items":
		docs, warnings, err := datasync.ParseItemsCSV(body)
		if err != nil {
			return diff, err
		}
		plan, err := datasync.PlanItems(ctx, q, docs)
		if err != nil {
			return diff, err
		}
		plan.Warnings = append(plan.Warnings, warnings...)
		for _, ip := range plan.Items {
			diff.Entries = append(diff.Entries, pages.DiffEntry{
				Kind: "item", Name: ip.Doc.Name, Action: string(ip.Action), Changes: ip.Changes, Nested: false,
			})
		}
		diff.Warnings = plan.Warnings
		diff.Counts = countsToStrings(plan.Counts)
	}
	return diff, nil
}

// applyDiff re-derives the plan for a source and applies it.
func (h *SyncHandler) applyDiff(ctx context.Context, src models.SyncSource) error {
	body, err := h.svc.Fetch(ctx, src)
	if err != nil {
		return err
	}
	defer body.Close()

	q := models.New(h.dbPool)
	switch src.Kind {
	case "roles":
		docs, _, err := datasync.ParseRolesCSV(body, models.Alignment(src.Alignment))
		if err != nil {
			return err
		}
		plan, err := datasync.PlanRoles(ctx, q, models.Alignment(src.Alignment), docs)
		if err != nil {
			return err
		}
		return datasync.ApplyRoles(ctx, h.dbPool, plan)
	case "items":
		docs, _, err := datasync.ParseItemsCSV(body)
		if err != nil {
			return err
		}
		plan, err := datasync.PlanItems(ctx, q, docs)
		if err != nil {
			return err
		}
		return datasync.ApplyItems(ctx, h.dbPool, plan)
	}
	return nil
}

// loadPageData assembles the sources + run history for the page.
func (h *SyncHandler) loadPageData(ctx context.Context) (pages.SyncPageData, error) {
	sources, err := h.svc.ListSources(ctx)
	if err != nil {
		return pages.SyncPageData{}, err
	}
	runs, err := h.svc.LastRuns(ctx, 15)
	if err != nil {
		return pages.SyncPageData{}, err
	}

	runViews := make([]pages.SyncRunView, len(runs))
	for i, r := range runs {
		counts := map[string]int{}
		_ = json.Unmarshal(r.ActionCounts, &counts)
		runViews[i] = pages.SyncRunView{
			SourceName: r.SourceName, Status: r.Status, Counts: counts,
			RunBy: r.RunBy, ErrorMessage: r.ErrorMessage, StartedAt: r.StartedAt.Time,
		}
	}

	// Latest run per source, for the sources card.
	latest := map[string]pages.SyncRunView{}
	for _, r := range runViews {
		if _, ok := latest[r.SourceName]; !ok {
			latest[r.SourceName] = r
		}
	}

	srcViews := make([]pages.SyncSourceView, len(sources))
	for i, s := range sources {
		srcViews[i] = pages.SyncSourceView{
			ID: s.ID, Name: s.Name, Kind: s.Kind, Alignment: s.Alignment,
			Url: s.Url, Enabled: s.Enabled,
		}
		if lr, ok := latest[s.Name]; ok {
			cp := lr
			srcViews[i].LastRun = &cp
		}
	}

	return pages.SyncPageData{Sources: srcViews, Runs: runViews}, nil
}

func countsToStrings(in map[datasync.Action]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[string(k)] = v
	}
	return out
}

// actionCounts converts the flattened string-keyed counts back to Action keys
// for sync_run recording.
func actionCounts(in map[string]int) map[datasync.Action]int {
	out := make(map[datasync.Action]int, len(in))
	for k, v := range in {
		out[datasync.Action(k)] = v
	}
	return out
}

func (h *SyncHandler) badRequest(c echo.Context, msg string) error {
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "`+msg+`", "type": "error"}}`)
	return c.String(http.StatusBadRequest, msg)
}

func (h *SyncHandler) syncError(c echo.Context, msg string) error {
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "`+msg+`", "type": "error"}}`)
	return c.String(http.StatusInternalServerError, msg)
}
