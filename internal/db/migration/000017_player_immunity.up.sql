CREATE TABLE IF NOT EXISTS player_immunity (
  player_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  status_id INT NOT NULL,
  one_time BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (player_id, status_id)
);

