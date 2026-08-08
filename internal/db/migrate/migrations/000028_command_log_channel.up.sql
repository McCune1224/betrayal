CREATE TABLE IF NOT EXISTS command_log_channel (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    channel_id VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
