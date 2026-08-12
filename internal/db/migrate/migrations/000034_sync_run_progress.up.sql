ALTER TABLE sync_run
    ADD COLUMN phase TEXT NOT NULL DEFAULT '',
    ADD COLUMN progress INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN total INTEGER NOT NULL DEFAULT 0;

UPDATE sync_run
SET phase = CASE status WHEN 'applied' THEN 'complete' WHEN 'failed' THEN 'failed' ELSE '' END,
    progress = CASE WHEN status IN ('applied', 'failed') THEN 1 ELSE 0 END,
    total = 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_sync_run_active_source
    ON sync_run (source_id)
    WHERE status IN ('pending', 'running');
