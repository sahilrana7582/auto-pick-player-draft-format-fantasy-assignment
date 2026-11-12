-- league_picks: whose turn it is right now in a league's draft.
-- One row per league (the "current turn" pointer).
-- When expires_at passes and the user still hasn't picked,
-- the auto-pick worker picks a player for that entry and moves the turn on.
CREATE TABLE IF NOT EXISTS league_picks (
    id          BIGSERIAL PRIMARY KEY,
    league_id   BIGINT NOT NULL UNIQUE REFERENCES leagues(id) ON DELETE CASCADE,
    entry_id    BIGINT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,  -- entry whose turn it is
    expires_at  TIMESTAMPTZ NOT NULL,                                      -- deadline; after this, auto-pick
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()                         -- when the turn last changed
);

-- the auto-pick worker scans for expired turns: WHERE expires_at <= NOW()
CREATE INDEX IF NOT EXISTS idx_league_picks_expires_at ON league_picks(expires_at);
