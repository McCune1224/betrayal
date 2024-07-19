CREATE TABLE IF NOT EXISTS player_ability (
  player_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  ability_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  PRIMARY KEY (player_id, ability_id)
);

