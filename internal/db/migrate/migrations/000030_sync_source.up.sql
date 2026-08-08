CREATE TABLE IF NOT EXISTS sync_source (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,          -- 'good_roles' | 'evil_roles' | 'neutral_roles' | 'items'
    kind       TEXT NOT NULL,                 -- 'roles' | 'items'
    alignment  TEXT NOT NULL DEFAULT '',      -- GOOD/EVIL/NEUTRAL or '' for items
    url        TEXT NOT NULL DEFAULT '',     -- real URL seeded from env at app startup
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed the four canonical sources so the web panel has rows before the first
-- run. URLs come from env at app startup when the row is missing (see
-- internal/services/datasync); the placeholder below is only a schema default.
INSERT INTO sync_source (name, kind, alignment) VALUES
    ('good_roles',   'roles', 'GOOD'),
    ('evil_roles',   'roles', 'EVIL'),
    ('neutral_roles','roles', 'NEUTRAL'),
    ('items',        'items', '')
ON CONFLICT (name) DO NOTHING;
