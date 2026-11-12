-- roster_owners: which player belongs to which entry (user) in which league.
-- One row = one picked player.
CREATE TABLE IF NOT EXISTS roster_owners (
    id          BIGSERIAL PRIMARY KEY,
    league_id   BIGINT NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    entry_id    BIGINT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),   -- when the player was picked

    -- draft rule: a player can be owned by only one entry per league.
    -- This also protects against two users picking the same player at the same time.
    CONSTRAINT uq_roster_league_player UNIQUE (league_id, player_id)
);

-- "show me this entry's team"
CREATE INDEX IF NOT EXISTS idx_roster_owners_entry_id ON roster_owners(entry_id);
