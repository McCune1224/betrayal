CREATE TABLE IF NOT EXISTS vote (
    id SERIAL PRIMARY KEY,
    voter_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
    target_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
    cycle_day INTEGER NOT NULL,
    is_elimination BOOLEAN NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    context TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    -- Unique constraint: one vote per voter per cycle (upsert pattern)
    UNIQUE(voter_id, cycle_day, is_elimination)
);

-- Index for querying votes by cycle
CREATE INDEX idx_vote_cycle ON vote(cycle_day, is_elimination);

-- Index for querying votes by target (for stats)
CREATE INDEX idx_vote_target ON vote(target_id);
