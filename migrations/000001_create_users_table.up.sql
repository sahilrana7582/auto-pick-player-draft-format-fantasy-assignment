-- users: a person who can join leagues.
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    age         INT CHECK (age > 0),                 -- basic sanity check
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
