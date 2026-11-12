package models

import "time"

// RosterOwner is the DB model for `roster_owners`: which entry owns which player in a league.
type RosterOwner struct {
	ID        int64
	LeagueID  int64
	EntryID   int64
	PlayerID  int64
	CreatedAt time.Time // when the player was picked

	Player Player // filled when the query JOINs players
}
