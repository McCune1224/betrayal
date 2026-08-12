-- name: ListSyncSources :many
SELECT * FROM sync_source ORDER BY id;

-- name: GetSyncSourceByName :one
SELECT * FROM sync_source WHERE name = $1;

-- name: UpdateSyncSource :one
UPDATE sync_source
SET url = $2, enabled = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpsertSyncSourceFromEnv :exec
-- Seeds the canonical sources at app startup. The URL from env is only used
-- when the row still has the empty placeholder (never overwrites a URL the
-- user edited in the web panel).
INSERT INTO sync_source (name, kind, alignment, url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (name) DO UPDATE SET
    url       = CASE WHEN sync_source.url = '' THEN EXCLUDED.url ELSE sync_source.url END,
    kind      = EXCLUDED.kind,
    alignment = EXCLUDED.alignment,
    updated_at = NOW();

-- name: CreateSyncRun :one
INSERT INTO sync_run (source_id, source_name, status, action_counts, run_by, error_message, started_at, finished_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: CreatePendingSyncRun :one
INSERT INTO sync_run (source_id, source_name, status, action_counts, run_by, error_message, started_at, phase, progress, total)
VALUES ($1, $2, 'pending', '{}', $3, '', NOW(), 'queued', 0, 4)
RETURNING *;

-- name: GetSyncRun :one
SELECT * FROM sync_run WHERE id = $1;

-- name: GetActiveSyncRunBySource :one
SELECT * FROM sync_run
WHERE source_id = $1 AND status IN ('pending', 'running')
ORDER BY started_at DESC
LIMIT 1;

-- name: UpdateSyncRun :one
UPDATE sync_run
SET status = $2, phase = $3, progress = $4, total = $5,
    action_counts = $6, error_message = $7, finished_at = $8
WHERE id = $1
RETURNING *;

-- name: ListSyncRuns :many
SELECT * FROM sync_run ORDER BY started_at DESC LIMIT $1;
