CREATE TABLE IF NOT EXISTS player_item (
  player_id INT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  item_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  PRIMARY KEY (player_id, item_id)
);
