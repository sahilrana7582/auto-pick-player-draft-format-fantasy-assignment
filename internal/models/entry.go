package models

import "time"

// Entry is the DB model for the `entries` table: one user's seat in one league.
type Entry struct {
	ID        int64
	LeagueID  int64
	UserID    int64
	PickOrder int // turn order in the draft: 1, 2, 3, ...
	JoinedAt  time.Time
	CreatedAt time.Time
}
