package dto

import (
	"time"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// PlayerResponse is what we send back to the client for a player.
type PlayerResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Position  string    `json:"position"` // Bat | Bowl | AR | Keeper
	CreatedAt time.Time `json:"created_at"`
}

// ToPlayerResponse converts a DB model into an API response.
func ToPlayerResponse(p models.Player) PlayerResponse {
	return PlayerResponse{
		ID:        p.ID,
		Name:      p.Name,
		Position:  string(p.Position),
		CreatedAt: p.CreatedAt,
	}
}

// ToPlayerResponses converts a slice of DB models.
func ToPlayerResponses(players []models.Player) []PlayerResponse {
	resp := make([]PlayerResponse, 0, len(players))
	for _, p := range players {
		resp = append(resp, ToPlayerResponse(p))
	}
	return resp
}
