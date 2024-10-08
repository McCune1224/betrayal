CREATE TABLE IF NOT EXISTS player_status (
  player_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  status_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (player_id, status_id)
);

