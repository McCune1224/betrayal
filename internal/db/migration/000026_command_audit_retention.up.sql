-- Add function to clean up command audits older than 365 days
CREATE OR REPLACE FUNCTION cleanup_old_command_audits() RETURNS void AS $$
BEGIN
    DELETE FROM command_audit 
    WHERE timestamp < NOW() - INTERVAL '365 days';
END;
$$ LANGUAGE plpgsql;

-- Create index on timestamp for faster deletion queries
CREATE INDEX IF NOT EXISTS idx_command_audit_timestamp ON command_audit(timestamp);

-- Add policy comment for documentation
COMMENT ON TABLE command_audit IS 'Command audit trail - automatically cleaned up after 365 days. See cleanup_old_command_audits() function.';
