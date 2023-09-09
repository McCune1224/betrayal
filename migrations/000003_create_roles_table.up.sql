CREATE TYPE alignment AS ENUM ('GOOD', 'NEUTRAL', 'EVIL');
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
    categories varchar[],
    charges integer NOT NULL default 0,
    any_ability boolean NOT NULL default false,
    rarity varchar NOT NULL default 'Role',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS perks (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS statuses (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name varchar NOT NULL UNIQUE,
    description TEXT NOT NULL,
    cost integer NOT NULL default 0,
    rarity varchar NOT NULL,
    categories varchar[] NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);


-- AHOY MATEY YE BE ENTERING THE DANGER ZONE, YE BE WARNED OF THE DANGER AHEAD OF MANY JOINTABLES AND FOREIGN KEYS

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



