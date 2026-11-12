package models

import "time"

// LeagueType is the format of a league (matches the CHECK constraint in the DB).
type LeagueType string

const (
	LeagueTypeSeasonLong LeagueType = "season_long" // normal fantasy, players can be shared
	LeagueTypeDraftPick  LeagueType = "draft_pick"  // turn-based draft, each player owned once
)

// IsValid reports whether t is one of the known league types.
func (t LeagueType) IsValid() bool {
	return t == LeagueTypeSeasonLong || t == LeagueTypeDraftPick
}

// LeagueStatus is the lifecycle state of a league.
//
//	created -> started -> completed
type LeagueStatus string

const (
	LeagueStatusCreated   LeagueStatus = "created"   // open: users can join
	LeagueStatusStarted   LeagueStatus = "started"   // draft running: joining is closed
	LeagueStatusCompleted LeagueStatus = "completed" // draft finished
)

// League is the DB model for the `leagues` table.
type League struct {
	ID          int64
	Name        string
	Description *string // nullable column -> pointer
	Type        LeagueType
	SeriesName  string
	Status      LeagueStatus
	CreatedAt   time.Time
}
