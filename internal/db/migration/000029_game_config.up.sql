CREATE TABLE IF NOT EXISTS game_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed defaults so `/inv create` behaves identically out of the box while
-- remaining editable. The application falls back to these same values when a
-- row is missing or unparseable.
INSERT INTO game_config (key, value) VALUES
    ('default_coins', '200'),
    ('default_items_limit', '4'),
    ('default_luck', '0')
ON CONFLICT (key) DO NOTHING;
