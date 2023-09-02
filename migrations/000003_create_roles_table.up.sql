CREATE TYPE alignment AS ENUM ('GOOD', 'NEUTRAL', 'EVIL');
CREATE TYPE rarity AS ENUM ('COMMON', 'UNCOMMON', 'RARE', 'EPIC', 'LEGENDARY', 'MYTHICAL', 'UNIQUE', 'ULTIMATE');
CREATE TYPE action_type AS ENUM ('POSITIVE', 'NEUTRAL', 'NEGATIVE');


CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL,
    alignment alignment NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS abilities (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL, 
    action_type varchar NOT null,
    any_ability boolean NOT NULL default false,
    rarity varchar NOT NULL default 'Role',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS perks (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL,
    rarity varchar NOT NULL default 'role perk',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS status (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles_abilities (
    id SERIAL PRIMARY KEY,
    role_id integer NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    ability_id integer NOT NULL REFERENCES abilities(id) ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS roles_perks (
    id SERIAL PRIMARY KEY,
    role_id integer NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    perk_id integer NOT NULL REFERENCES perks(id) ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);


