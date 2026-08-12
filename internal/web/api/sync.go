package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/logger"
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
	pool         *pgxpool.Pool
	service      *datasync.Service
	jobs         chan syncJob
	stop         chan struct{}
	done         chan struct{}
	workerCtx    context.Context
	cancelWorker context.CancelFunc
	mu           sync.Mutex
	running      map[int32]bool
}

type syncJob struct {
	runID  int64
	source models.SyncSource
}

type syncPhaseError struct {
	phase string
	err   error
}

func (e *syncPhaseError) Error() string { return fmt.Sprintf("%s: %v", e.phase, e.err) }
func (e *syncPhaseError) Unwrap() error { return e.err }

func wrapSyncPhase(phase string, err error) error {
	if err == nil {
		return nil
	}
	return &syncPhaseError{phase: phase, err: err}
}

func syncPhase(err error) string {
	var phaseErr *syncPhaseError
	if errors.As(err, &phaseErr) {
		return phaseErr.phase
	}
	return "unknown"
}

func NewSyncHandler(pool *pgxpool.Pool, service *datasync.Service) *SyncHandler {
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	h := &SyncHandler{pool: pool, service: service, jobs: make(chan syncJob, 8), stop: make(chan struct{}), done: make(chan struct{}), workerCtx: workerCtx, cancelWorker: cancelWorker, running: make(map[int32]bool)}
	go h.worker()
	return h
}

func (h *SyncHandler) Shutdown() {
	select {
	case <-h.stop:
		return
	default:
		close(h.stop)
	}
	h.cancelWorker()
	<-h.done
}

func (h *SyncHandler) worker() {
	defer close(h.done)
	for {
		select {
		case job := <-h.jobs:
			h.process(job)
		case <-h.stop:
			return
		}
	}
}

func (h *SyncHandler) process(job syncJob) {
	ctx, cancel := context.WithTimeout(h.workerCtx, 120*time.Second)
	defer cancel()
	h.updateRun(ctx, job.runID, "running", "starting", 0, 4, nil, "")
	counts, err := h.applySource(ctx, job.source, func(phase string, progress int32) {
		h.updateRun(ctx, job.runID, "running", phase, progress, 4, nil, "")
	})
	h.mu.Lock()
	delete(h.running, job.source.ID)
	h.mu.Unlock()
	if err != nil {
		phase := syncPhase(err)
		h.updateRun(ctx, job.runID, datasync.RunStatusFailed, phase, 4, 4, nil, err.Error())
		logger.Get().Error().Err(err).Str("operation", "sync_apply").Str("source", job.source.Name).Str("phase", phase).Msg("catalog sync apply failed")
		return
	}
	h.updateRun(ctx, job.runID, datasync.RunStatusApplied, "complete", 4, 4, counts, "")
}

func (h *SyncHandler) updateRun(ctx context.Context, id int64, status, phase string, progress, total int32, counts map[datasync.Action]int, message string) {
	data, _ := json.Marshal(actionCounts(counts))
	finishedAt := pgtype.Timestamptz{}
	if status == datasync.RunStatusApplied || status == datasync.RunStatusFailed {
		finishedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}
	_, err := models.New(h.pool).UpdateSyncRun(ctx, models.UpdateSyncRunParams{ID: id, Status: status, Phase: phase, Progress: progress, Total: total, ActionCounts: data, ErrorMessage: message, FinishedAt: finishedAt})
	if err != nil {
		logger.Get().Error().Err(err).Int64("run_id", id).Msg("could not update sync run")
	}
}

func syncRunDTO(run models.SyncRun) map[string]any {
	counts := map[string]int{}
	_ = json.Unmarshal(run.ActionCounts, &counts)
	return map[string]any{"id": run.ID, "source_id": run.SourceID, "source_name": run.SourceName, "status": run.Status, "phase": run.Phase, "progress": run.Progress, "total": run.Total, "counts": counts, "error": run.ErrorMessage, "started_at": run.StartedAt, "finished_at": run.FinishedAt}
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
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()
	sources, err := h.service.ListSources(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "sync_unavailable", "could not load sync sources", nil)
		return nil
	}
	previews := make([]map[string]any, 0, len(sources))
	for _, source := range sources {
		entry := map[string]any{"source": sourceDTO(source), "enabled": source.Enabled}
		if !source.Enabled || strings.TrimSpace(source.Url) == "" {
			entry["status"] = "skipped"
			entry["reason"] = "source is disabled or has no URL"
			previews = append(previews, entry)
			continue
		}
		plan, counts, err := h.planSource(ctx, source)
		if err != nil {
			entry["status"] = "failed"
			entry["error"] = err.Error()
		} else {
			entry["status"] = "ready"
			entry["counts"] = actionCounts(counts)
			entry["plan"] = plan
		}
		previews = append(previews, entry)
	}
	WriteJSON(c.Response(), 200, map[string]any{"previews": previews, "read_only": true, "status": "preview_ready"})
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
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
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
	h.mu.Lock()
	if h.running[source.ID] {
		h.mu.Unlock()
		WriteError(c.Response(), 409, "sync_already_running", "a sync for this source is already running", nil)
		return nil
	}
	h.running[source.ID] = true
	h.mu.Unlock()
	active, activeErr := models.New(h.pool).GetActiveSyncRunBySource(ctx, pgtype.Int4{Int32: source.ID, Valid: true})
	if activeErr == nil && active.ID != 0 {
		h.mu.Lock()
		delete(h.running, source.ID)
		h.mu.Unlock()
		WriteJSON(c.Response(), http.StatusAccepted, map[string]any{"run": syncRunDTO(active), "status": "already_running"})
		return nil
	}
	run, err := models.New(h.pool).CreatePendingSyncRun(ctx, models.CreatePendingSyncRunParams{
		SourceID: pgtype.Int4{Int32: source.ID, Valid: true}, SourceName: source.Name, RunBy: "web",
	})
	if err != nil {
		h.mu.Lock()
		delete(h.running, source.ID)
		h.mu.Unlock()
		logger.Get().Error().Err(err).Str("operation", "sync_queue").Str("source", source.Name).Msg("could not create sync run")
		WriteError(c.Response(), 500, "sync_queue_failed", "could not queue sync", nil)
		return nil
	}
	select {
	case h.jobs <- syncJob{runID: run.ID, source: source}:
		WriteJSON(c.Response(), http.StatusAccepted, map[string]any{"run": syncRunDTO(run), "status": "queued"})
	default:
		h.mu.Lock()
		delete(h.running, source.ID)
		h.mu.Unlock()
		h.updateRun(context.Background(), run.ID, datasync.RunStatusFailed, "queue", 0, 4, nil, "sync worker queue is full")
		WriteError(c.Response(), http.StatusServiceUnavailable, "sync_queue_full", "sync queue is full; try again shortly", nil)
	}
	return nil
}

func (h *SyncHandler) Run(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(c.Response(), 400, "invalid_run", "run id must be a positive integer", nil)
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()
	run, err := models.New(h.pool).GetSyncRun(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "run_not_found", "sync run not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, map[string]any{"run": syncRunDTO(run)})
	return nil
}

func (h *SyncHandler) planSource(ctx context.Context, source models.SyncSource) (any, map[datasync.Action]int, error) {
	body, err := h.service.Fetch(ctx, source)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()
	q := models.New(h.pool)
	switch source.Kind {
	case "roles":
		docs, _, err := datasync.ParseRolesCSV(body, models.Alignment(source.Alignment))
		if err != nil {
			return nil, nil, err
		}
		plan, err := datasync.PlanRoles(ctx, q, models.Alignment(source.Alignment), docs)
		if err != nil {
			return nil, nil, err
		}
		return plan, plan.Counts, nil
	case "items":
		docs, _, err := datasync.ParseItemsCSV(body)
		if err != nil {
			return nil, nil, err
		}
		plan, err := datasync.PlanItems(ctx, q, docs)
		if err != nil {
			return nil, nil, err
		}
		return plan, plan.Counts, nil
	default:
		return nil, nil, errors.New("unsupported sync source kind")
	}
}

func (h *SyncHandler) applySource(ctx context.Context, source models.SyncSource, progress func(string, int32)) (map[datasync.Action]int, error) {
	progress("fetch", 1)
	body, err := h.service.Fetch(ctx, source)
	if err != nil {
		return nil, wrapSyncPhase("fetch", err)
	}
	defer body.Close()
	q := models.New(h.pool)
	switch source.Kind {
	case "roles":
		progress("parse", 2)
		docs, _, err := datasync.ParseRolesCSV(body, models.Alignment(source.Alignment))
		if err != nil {
			return nil, wrapSyncPhase("parse", err)
		}
		plan, err := datasync.PlanRoles(ctx, q, models.Alignment(source.Alignment), docs)
		if err != nil {
			return nil, wrapSyncPhase("plan", err)
		}
		progress("apply", 3)
		if err := datasync.ApplyRoles(ctx, h.pool, plan); err != nil {
			return nil, wrapSyncPhase("apply", err)
		}
		return plan.Counts, nil
	case "items":
		progress("parse", 2)
		docs, _, err := datasync.ParseItemsCSV(body)
		if err != nil {
			return nil, wrapSyncPhase("parse", err)
		}
		plan, err := datasync.PlanItems(ctx, q, docs)
		if err != nil {
			return nil, wrapSyncPhase("plan", err)
		}
		progress("apply", 3)
		if err := datasync.ApplyItems(ctx, h.pool, plan); err != nil {
			return nil, wrapSyncPhase("apply", err)
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
