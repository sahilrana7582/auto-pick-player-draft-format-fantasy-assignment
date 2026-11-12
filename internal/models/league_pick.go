package models

import "time"

// LeaguePick is the DB model for `league_picks`: whose turn it is in a league.
type LeaguePick struct {
	ID        int64
	LeagueID  int64
	EntryID   int64     // entry whose turn it is
	ExpiresAt time.Time // after this, auto-pick for EntryID
	UpdatedAt time.Time
}
