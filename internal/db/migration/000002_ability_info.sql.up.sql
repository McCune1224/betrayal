CREATE TABLE IF NOT EXISTS ability_info (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  default_charges INTEGER NOT NULL,
  any_ability BOOLEAN NOT NULL,
  rarity rarity NOT NULL
);
