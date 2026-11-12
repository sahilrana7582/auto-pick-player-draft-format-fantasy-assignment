-- entries: a user's participation in a league.
-- pick_order is the turn order inside the league (1, 2, 3, ...),
-- assigned incrementally as users join.
-- ("order" is a reserved SQL keyword, so the column is named pick_order.)
CREATE TABLE IF NOT EXISTS entries (
    id          BIGSERIAL PRIMARY KEY,
    league_id   BIGINT NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    pick_order  INT    NOT NULL CHECK (pick_order > 0),
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),   -- when the user joined the league
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),   -- when the row was created

    -- a user can join the same league only once
    CONSTRAINT uq_entries_league_user  UNIQUE (league_id, user_id),
    -- two entries in the same league can't share the same turn number
    CONSTRAINT uq_entries_league_order UNIQUE (league_id, pick_order)
);

-- "which leagues has this user joined?"
CREATE INDEX IF NOT EXISTS idx_entries_user_id ON entries(user_id);
