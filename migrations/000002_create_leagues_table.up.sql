-- leagues: a contest users join.
-- type decides the format:
--   season_long -> normal fantasy (players can be shared)
--   draft_pick  -> snake/turn based draft (each player owned by one entry)
CREATE TABLE IF NOT EXISTS leagues (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(150) NOT NULL,
    description  TEXT,
    type         VARCHAR(20)  NOT NULL CHECK (type IN ('season_long', 'draft_pick')),
    series_name  VARCHAR(150) NOT NULL,              -- e.g. "IPL 2026"
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
