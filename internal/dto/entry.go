package dto

import (
	"time"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// CreateEntryRequest is the JSON body for POST /leagues/{id}/entries (join a league).
// league_id comes from the URL; pick_order is assigned by the server.
type CreateEntryRequest struct {
	UserID int64 `json:"user_id"`
}

// UpdateEntryRequest is the JSON body for PUT /entries/{id}.
// Only the draft order can change; moving an entry to another league/user makes no sense.
type UpdateEntryRequest struct {
	PickOrder int `json:"pick_order"`
}

// EntryResponse is what we send back to the client.
type EntryResponse struct {
	ID        int64     `json:"id"`
	LeagueID  int64     `json:"league_id"`
	UserID    int64     `json:"user_id"`
	PickOrder int       `json:"pick_order"`
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ToEntryResponse converts a DB model into an API response.
func ToEntryResponse(e models.Entry) EntryResponse {
	return EntryResponse{
		ID:        e.ID,
		LeagueID:  e.LeagueID,
		UserID:    e.UserID,
		PickOrder: e.PickOrder,
		JoinedAt:  e.JoinedAt,
		CreatedAt: e.CreatedAt,
	}
}

// ToEntryResponses converts a slice of DB models.
func ToEntryResponses(entries []models.Entry) []EntryResponse {
	resp := make([]EntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, ToEntryResponse(e))
	}
	return resp
}
