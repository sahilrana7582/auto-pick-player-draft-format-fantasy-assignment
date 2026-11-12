package dto

import (
	"time"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// RosterPlayerResponse is one player on an entry's roster.
type RosterPlayerResponse struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Position string    `json:"position"`
	PickedAt time.Time `json:"picked_at"`
}

// EntryRosterResponse is one entry (team) in a league with all its players.
type EntryRosterResponse struct {
	EntryID   int64                  `json:"entry_id"`
	UserID    int64                  `json:"user_id"`
	PickOrder int                    `json:"pick_order"`
	Players   []RosterPlayerResponse `json:"players"` // [] when nothing picked yet
}

// ToRosterPlayerResponse converts a roster row (with its joined Player) into a response.
func ToRosterPlayerResponse(r models.RosterOwner) RosterPlayerResponse {
	return RosterPlayerResponse{
		ID:       r.Player.ID,
		Name:     r.Player.Name,
		Position: string(r.Player.Position),
		PickedAt: r.CreatedAt,
	}
}
