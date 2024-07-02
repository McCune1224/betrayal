CREATE TABLE IF NOT EXISTS player_perk (
  player_id INT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  perk_id INT NOT NULL,
  PRIMARY KEY (player_id, perk_id)
);

