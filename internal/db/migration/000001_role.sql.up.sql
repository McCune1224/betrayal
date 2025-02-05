CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;
CREATE TYPE alignment AS ENUM ('GOOD', 'NEUTRAL', 'EVIL');
CREATE TYPE rarity AS ENUM ('COMMON', 'UNCOMMON', 'RARE', 'EPIC', 'LEGENDARY', 'MYTHICAL', 'ROLE_SPECIFIC', 'UNIQUE');

CREATE TABLE IF NOT EXISTS role (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL,
  alignment alignment NOT NULL
);
