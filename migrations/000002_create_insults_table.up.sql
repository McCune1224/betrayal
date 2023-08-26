CREATE TABLE IF NOT EXISTS insults (
    id SERIAL PRIMARY KEY,
    author_id TEXT NOT NULL UNIQUE,
    insult TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
