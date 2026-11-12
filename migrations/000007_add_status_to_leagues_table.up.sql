-- League lifecycle:
--   created   -> users can join (default for new leagues)
--   started   -> draft in progress; no new entries allowed
--   completed -> draft finished
-- Existing rows get 'created' via the DEFAULT.
ALTER TABLE leagues
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'created'
        CONSTRAINT chk_leagues_status CHECK (status IN ('created', 'started', 'completed'));
