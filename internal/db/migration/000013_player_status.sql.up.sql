CREATE TABLE IF NOT EXISTS player_status (
  player_id INT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  status_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  PRIMARY KEY (player_id, status_id)
);

