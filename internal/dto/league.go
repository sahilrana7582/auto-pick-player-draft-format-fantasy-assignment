package dto

import (
	"time"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// CreateLeagueRequest is the JSON body for POST /leagues.
type CreateLeagueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"` // optional
	Type        string `json:"type"`        // "season_long" | "draft_pick"
	SeriesName  string `json:"series_name"`
}

// UpdateLeagueRequest is the JSON body for PUT /leagues/{id} (replaces all fields).
// Status is NOT editable here; use POST /leagues/{id}/start.
type UpdateLeagueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	SeriesName  string `json:"series_name"`
}

// LeagueResponse is what we send back to the client.
type LeagueResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	SeriesName  string    `json:"series_name"`
	Status      string    `json:"status"` // created | started | completed
	CreatedAt   time.Time `json:"created_at"`
}

// ToLeagueResponse converts a DB model into an API response.
// A NULL description is returned as "".
func ToLeagueResponse(l models.League) LeagueResponse {
	desc := ""
	if l.Description != nil {
		desc = *l.Description
	}
	return LeagueResponse{
		ID:          l.ID,
		Name:        l.Name,
		Description: desc,
		Type:        string(l.Type),
		SeriesName:  l.SeriesName,
		Status:      string(l.Status),
		CreatedAt:   l.CreatedAt,
	}
}
