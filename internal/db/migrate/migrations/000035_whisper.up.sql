CREATE TABLE whisper_group (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE whisper_group_member (
    group_id BIGINT NOT NULL REFERENCES whisper_group(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES player(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, player_id),
    UNIQUE (player_id)
);

CREATE TABLE whisper_doubt_message (
    id BIGSERIAL PRIMARY KEY,
    message TEXT NOT NULL CHECK (char_length(btrim(message)) BETWEEN 1 AND 1000),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

INSERT INTO whisper_doubt_message (message) VALUES
    ('Do not trust everything you hear. Watch your back.'),
    ('That story may not be the whole truth. Keep your guard up.'),
    ('Someone may be trying to steer you. Verify what you can.'),
    ('Be careful who you rely on. Not everyone has your interests at heart.'),
    ('Pay attention to what is left unsaid.'),
    ('You may want to question that information before acting on it.');
