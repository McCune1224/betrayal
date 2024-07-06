CREATE TABLE IF NOT EXISTS player_role (
  player_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  role_id INT NOT NULL,
  PRIMARY KEY (player_id, role_id)
);

