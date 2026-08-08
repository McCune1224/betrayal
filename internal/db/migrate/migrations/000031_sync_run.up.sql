CREATE TABLE IF NOT EXISTS sync_run (
    id            BIGSERIAL PRIMARY KEY,
    source_id     INTEGER REFERENCES sync_source(id) ON DELETE CASCADE,
    source_name   TEXT NOT NULL,
    status        TEXT NOT NULL,               -- 'preview' | 'applied' | 'failed'
    action_counts JSONB NOT NULL DEFAULT '{}', -- {"created":3,"updated":5,"skipped":40}
    run_by        TEXT NOT NULL DEFAULT '',    -- web session user or 'cli'
    error_message TEXT NOT NULL DEFAULT '',
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sync_run_source_started
    ON sync_run (source_name, started_at DESC);
