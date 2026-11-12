-- leagues: a contest users join.
CREATE TABLE IF NOT EXISTS leagues (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(150) NOT NULL,
    description  TEXT,

    -- type decides the format:
    --   season_long -> normal fantasy (players can be shared)
    --   draft_pick  -> turn based draft (each player owned by one entry)
    type         VARCHAR(20)  NOT NULL CHECK (type IN ('season_long', 'draft_pick')),

    series_name  VARCHAR(150) NOT NULL,              -- e.g. "IPL 2026"

    -- lifecycle:
    --   created   -> users can join (default)
    --   started   -> draft in progress; no new entries allowed
    --   completed -> draft finished
    status       VARCHAR(20)  NOT NULL DEFAULT 'created'
                 CHECK (status IN ('created', 'started', 'completed')),

    -- max number of entries (members) that can join; the league can only
    -- start once exactly this many have joined
    size         INT          NOT NULL CHECK (size >= 2),

    -- seconds each entry gets to make a pick before auto-pick kicks in (1..300)
    timeout      INT          NOT NULL CHECK (timeout BETWEEN 1 AND 300),

    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
