package models

import "time"

// PlayerPosition is a player's role (matches the CHECK constraint in the DB).
type PlayerPosition string

const (
	PositionBat    PlayerPosition = "Bat"    // batter
	PositionBowl   PlayerPosition = "Bowl"   // bowler
	PositionAR     PlayerPosition = "AR"     // all-rounder
	PositionKeeper PlayerPosition = "Keeper" // wicket-keeper
)

// IsValid reports whether p is one of the known positions.
func (p PlayerPosition) IsValid() bool {
	switch p {
	case PositionBat, PositionBowl, PositionAR, PositionKeeper:
		return true
	}
	return false
}

// Player is the DB model for the `players` table.
type Player struct {
	ID        int64
	Name      string
	Position  PlayerPosition
	CreatedAt time.Time
}
