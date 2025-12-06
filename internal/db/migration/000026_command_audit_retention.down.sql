-- Drop the cleanup function and index
DROP FUNCTION IF EXISTS cleanup_old_command_audits();
DROP INDEX IF EXISTS idx_command_audit_timestamp;
