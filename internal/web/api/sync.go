package api

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
)

type SyncSourceDTO struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Alignment string `json:"alignment"`
	URL       string `json:"url"`
	Enabled   bool   `json:"enabled"`
}
type SyncDTO struct {
	Sources []SyncSourceDTO `json:"sources"`
}
type SyncHandler struct {
	pool    *pgxpool.Pool
	service *datasync.Service
}

func NewSyncHandler(pool *pgxpool.Pool, service *datasync.Service) *SyncHandler {
	return &SyncHandler{pool: pool, service: service}
}
func sourceDTO(source models.SyncSource) SyncSourceDTO {
	return SyncSourceDTO{ID: source.ID, Name: source.Name, Kind: source.Kind, Alignment: string(source.Alignment), URL: source.Url, Enabled: source.Enabled}
}

func (h *SyncHandler) Sources(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	sources, err := h.service.ListSources(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "sync_unavailable", "could not load sync sources", nil)
		return nil
	}
	out := make([]SyncSourceDTO, 0, len(sources))
	for _, s := range sources {
		out = append(out, SyncSourceDTO{ID: s.ID, Name: s.Name, Kind: s.Kind, Alignment: string(s.Alignment), URL: s.Url, Enabled: s.Enabled})
	}
	WriteJSON(c.Response(), 200, SyncDTO{Sources: out})
	return nil
}
func (h *SyncHandler) UpdateSource(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		WriteError(c.Response(), 400, "invalid_source", "source id must be an integer", nil)
		return nil
	}
	var req struct {
		URL     string `json:"url"`
		Enabled bool   `json:"enabled"`
	}
	if err := c.Bind(&req); err != nil {
		WriteError(c.Response(), 400, "invalid_request", "invalid JSON body", nil)
		return nil
	}
	req.URL = strings.TrimSpace(req.URL)
	if req.Enabled && req.URL == "" {
		WriteError(c.Response(), 400, "validation_error", "url is required for an enabled source", map[string]any{"url": "required"})
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	source, err := h.service.UpdateSource(ctx, int32(id), req.URL, req.Enabled)
	if err != nil {
		WriteError(c.Response(), 400, "source_update_failed", "could not update source", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, SyncSourceDTO{ID: source.ID, Name: source.Name, Kind: source.Kind, Alignment: string(source.Alignment), URL: source.Url, Enabled: source.Enabled})
	return nil
}
func (h *SyncHandler) Preview(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	sources, err := h.service.ListSources(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "sync_unavailable", "could not load sync sources", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, map[string]any{"sources": sources, "read_only": true, "status": "preview_ready"})
	return nil
}
func (h *SyncHandler) Apply(c echo.Context) error {
	var req struct {
		SourceID int32 `json:"source_id"`
	}
	if err := c.Bind(&req); err != nil || req.SourceID <= 0 {
		WriteError(c.Response(), 400, "invalid_source", "source_id is required", nil)
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 90*time.Second)
	defer cancel()
	sources, err := h.service.ListSources(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "sync_unavailable", "could not load sync sources", nil)
		return nil
	}
	var source models.SyncSource
	for _, candidate := range sources {
		if candidate.ID == req.SourceID {
			source = candidate
			break
		}
	}
	if source.ID == 0 {
		WriteError(c.Response(), 400, "invalid_source", "source was not found", nil)
		return nil
	}
	if !source.Enabled || strings.TrimSpace(source.Url) == "" {
		WriteError(c.Response(), 400, "invalid_source", "source is disabled or has no URL", nil)
		return nil
	}
	counts, err := h.applySource(ctx, source)
	if err != nil {
		_ = h.service.RecordRun(ctx, &source.ID, source.Name, datasync.RunStatusFailed, "web", err.Error(), nil)
		WriteError(c.Response(), 502, "sync_apply_failed", "sync apply failed", nil)
		return nil
	}
	_ = h.service.RecordRun(ctx, &source.ID, source.Name, datasync.RunStatusApplied, "web", "", counts)
	WriteJSON(c.Response(), 200, map[string]any{"source": sourceDTO(source), "status": "applied", "counts": actionCounts(counts)})
	return nil
}

func (h *SyncHandler) applySource(ctx context.Context, source models.SyncSource) (map[datasync.Action]int, error) {
	body, err := h.service.Fetch(ctx, source)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	q := models.New(h.pool)
	switch source.Kind {
	case "roles":
		docs, _, err := datasync.ParseRolesCSV(body, models.Alignment(source.Alignment))
		if err != nil {
			return nil, err
		}
		plan, err := datasync.PlanRoles(ctx, q, models.Alignment(source.Alignment), docs)
		if err != nil {
			return nil, err
		}
		if err := datasync.ApplyRoles(ctx, h.pool, plan); err != nil {
			return nil, err
		}
		return plan.Counts, nil
	case "items":
		docs, _, err := datasync.ParseItemsCSV(body)
		if err != nil {
			return nil, err
		}
		plan, err := datasync.PlanItems(ctx, q, docs)
		if err != nil {
			return nil, err
		}
		if err := datasync.ApplyItems(ctx, h.pool, plan); err != nil {
			return nil, err
		}
		return plan.Counts, nil
	default:
		return nil, errors.New("unsupported sync source kind")
	}
}

func actionCounts(in map[datasync.Action]int) map[string]int {
	out := make(map[string]int, len(in))
	for action, count := range in {
		out[string(action)] = count
	}
	return out
}
