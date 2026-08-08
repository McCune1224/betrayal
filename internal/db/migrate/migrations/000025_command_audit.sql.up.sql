-- Create command_audit table to track all player/admin commands with their arguments
CREATE TABLE IF NOT EXISTS command_audit (
    id BIGSERIAL PRIMARY KEY,
    correlation_id UUID,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    command_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL,
    user_roles TEXT[] DEFAULT ARRAY[]::TEXT[],
    guild_id TEXT,
    channel_id TEXT,
    is_admin BOOLEAN DEFAULT FALSE,
    command_arguments JSONB,
    status TEXT DEFAULT 'success', -- 'success', 'error', 'cancelled'
    error_message TEXT,
    execution_time_ms INTEGER,
    environment TEXT DEFAULT 'local'
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_command_audit_timestamp ON command_audit(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_command_audit_user_id ON command_audit(user_id);
CREATE INDEX IF NOT EXISTS idx_command_audit_command_name ON command_audit(command_name);
CREATE INDEX IF NOT EXISTS idx_command_audit_correlation_id ON command_audit(correlation_id);
CREATE INDEX IF NOT EXISTS idx_command_audit_is_admin ON command_audit(is_admin);
CREATE INDEX IF NOT EXISTS idx_command_audit_status ON command_audit(status);
CREATE INDEX IF NOT EXISTS idx_command_audit_guild_id ON command_audit(guild_id);
