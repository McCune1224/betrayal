-- Unique names across the catalog. The spreadsheet sync (internal/services/
-- datasync) matches by exact name; duplicates would make that ambiguous.
-- If this migration fails on an existing database, there are duplicate rows
-- that must be cleaned up manually first (the migration UI will show it as a
-- failed/dirty migration).
CREATE UNIQUE INDEX IF NOT EXISTS uq_role_name ON role (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_item_name ON item (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_ability_info_name ON ability_info (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_perk_info_name ON perk_info (name);
