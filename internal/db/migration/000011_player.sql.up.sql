CREATE TABLE IF NOT EXISTS player (
  id BIGINT PRIMARY KEY, -- this is the discord user id technically...
  role_id INT references role(id) ON DELETE CASCADE,
  alive BOOLEAN NOT NULL DEFAULT TRUE,
  coins INT NOT NULL DEFAULT 200,
  luck INT NOT NULL DEFAULT 0,
  item_limit INT NOT NULL DEFAULT 4,
  alignment alignment NOT NULL
)

