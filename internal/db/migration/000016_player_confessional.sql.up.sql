CREATE TABLE IF NOT EXISTS player_confessional (
  player_id INT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  channel_id INT NOT NULL,
  pin_message_id INT NOT NULL,
  PRIMARY KEY (player_id, channel_id)
);
