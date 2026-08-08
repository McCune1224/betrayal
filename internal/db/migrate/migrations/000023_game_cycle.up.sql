CREATE TABLE IF NOT EXISTS game_cycle (
    id SERIAL PRIMARY KEY,
    is_elimination BOOLEAN NOT NULL DEFAULT FALSE,
    day integer NOT NULL DEFAULT 0
);

-- Day 0 is game start (transition normally goes Day n -> Elimination n OR Elimination n -> Day n+1)
-- Day 0 will be an edge case however where it goes from Day 0 -> Day 1
INSERT INTO game_cycle (is_elimination, day) VALUES (FALSE, 0) RETURNING *;

