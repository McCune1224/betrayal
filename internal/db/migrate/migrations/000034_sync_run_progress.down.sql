ALTER TABLE sync_run
    DROP COLUMN IF EXISTS phase,
    DROP COLUMN IF EXISTS progress,
    DROP COLUMN IF EXISTS total;

DROP INDEX IF EXISTS uq_sync_run_active_source;
