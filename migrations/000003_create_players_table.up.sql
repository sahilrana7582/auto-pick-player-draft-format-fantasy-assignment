-- players: the real-world cricket players that can be picked.
CREATE TABLE IF NOT EXISTS players (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(150) NOT NULL,
    position    VARCHAR(10)  NOT NULL CHECK (position IN ('Bat', 'Bowl', 'AR', 'Keeper')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
