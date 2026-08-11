package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/services/datasync"
	"strconv"
	"strings"
	"time"
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
type SyncHandler struct{ service *datasync.Service }

func NewSyncHandler(service *datasync.Service) *SyncHandler { return &SyncHandler{service: service} }
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
	WriteError(c.Response(), 501, "not_implemented", "sync apply API is not yet available", nil)
	return nil
}
