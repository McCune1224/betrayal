package datasync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/models"
)

// EnvURLsFromEnv builds the source-name → CSV URL map from the environment
// (GOOD_ROLES_CSV / EVIL_ROLES_CSV / NEUTRAL_ROLES_CSV / ITEM_CSV). Used by
// the web server startup seeding and the archived CLI.
func EnvURLsFromEnv() map[string]string {
	out := make(map[string]string, len(envKeys))
	for name, key := range envKeys {
		out[name] = os.Getenv(key)
	}
	return out
}

// SeedSources upserts the four canonical sources from the environment. A URL
// is only written when the row still has the empty placeholder, so URLs the
// user edits in the web panel are never clobbered by a restart.
func (s *Service) SeedSources(ctx context.Context) error {
	q := models.New(s.pool)
	for _, src := range CanonicalSources {
		url := s.envURLs[src.Name]
		if err := q.UpsertSyncSourceFromEnv(ctx, models.UpsertSyncSourceFromEnvParams{
			Name:      src.Name,
			Kind:      src.Kind,
			Alignment: src.Alignment,
			Url:       url,
		}); err != nil {
			return fmt.Errorf("seed sync source %q: %w", src.Name, err)
		}
	}
	return nil
}

// ListSources returns all sync sources ordered by id.
func (s *Service) ListSources(ctx context.Context) ([]models.SyncSource, error) {
	return models.New(s.pool).ListSyncSources(ctx)
}

// UpdateSource updates a source's URL and enabled flag.
func (s *Service) UpdateSource(ctx context.Context, id int32, url string, enabled bool) (models.SyncSource, error) {
	return models.New(s.pool).UpdateSyncSource(ctx, models.UpdateSyncSourceParams{
		ID: id, Url: url, Enabled: enabled,
	})
}

// RunStatus values recorded in sync_run.
const (
	RunStatusPreview = "preview" // a fetch+diff completed without applying
	RunStatusApplied = "applied" // a plan was applied to the database
	RunStatusFailed  = "failed"  // fetch, parse, or apply errored
)

// RecordRun writes a sync_run audit row. Counts are stored as JSONB.
// sourceID may be nil for runs not tied to a source.
func (s *Service) RecordRun(ctx context.Context, sourceID *int32, sourceName, status, runBy, errMsg string, counts map[Action]int) error {
	countsJSON, err := json.Marshal(counts)
	if err != nil {
		return err
	}

	now := time.Now()
	var id pgtype.Int4
	if sourceID != nil {
		id = pgtype.Int4{Int32: *sourceID, Valid: true}
	}
	_, err = models.New(s.pool).CreateSyncRun(ctx, models.CreateSyncRunParams{
		SourceID:     id,
		SourceName:   sourceName,
		Status:       status,
		ActionCounts: countsJSON,
		RunBy:        runBy,
		ErrorMessage: errMsg,
		StartedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		FinishedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	})
	return err
}

// LastRuns returns the most recent sync runs for the history card.
func (s *Service) LastRuns(ctx context.Context, limit int32) ([]models.SyncRun, error) {
	return models.New(s.pool).ListSyncRuns(ctx, limit)
}
