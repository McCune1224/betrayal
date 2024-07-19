CREATE TABLE IF NOT EXISTS player_confessional (
  player_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  channel_id BIGINT NOT NULL,
  pin_message_id BIGINT NOT NULL,
  PRIMARY KEY (player_id, channel_id)
);

