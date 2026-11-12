package dto

import (
	"time"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// PickPlayerRequest is the JSON body for a manual pick by the entry on the clock.
// (league_id / entry_id come from the URL, not the body.)
type PickPlayerRequest struct {
	PlayerID int64 `json:"player_id"`
}

// LeaguePickResponse is the current turn in a league's draft.
type LeaguePickResponse struct {
	ID        int64     `json:"id"`
	LeagueID  int64     `json:"league_id"`
	EntryID   int64     `json:"entry_id"`   // entry whose turn it is
	ExpiresAt time.Time `json:"expires_at"` // after this, the worker auto-picks
	UpdatedAt time.Time `json:"updated_at"` // when the turn last moved
}

// ToLeaguePickResponse converts a DB model into an API response.
func ToLeaguePickResponse(p models.LeaguePick) LeaguePickResponse {
	return LeaguePickResponse{
		ID:        p.ID,
		LeagueID:  p.LeagueID,
		EntryID:   p.EntryID,
		ExpiresAt: p.ExpiresAt,
		UpdatedAt: p.UpdatedAt,
	}
}
