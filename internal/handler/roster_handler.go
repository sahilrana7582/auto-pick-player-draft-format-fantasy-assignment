package handler

import (
	"net/http"

	"auto-pick-player-draft-format-fantasy-assignment/internal/service"
)

// RosterHandler only deals with HTTP: parse input, call service, write output.
type RosterHandler struct {
	svc *service.RosterService
}

func NewRosterHandler(svc *service.RosterService) *RosterHandler {
	return &RosterHandler{svc: svc}
}

// RegisterRoutes wires the roster endpoints onto the mux.
func (h *RosterHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /leagues/{id}/rosters", h.ListByLeague)
	mux.HandleFunc("GET /entries/{id}/roster", h.GetByEntry)
}

// GET /leagues/{id}/rosters  -> every entry in draft order with its players
func (h *RosterHandler) ListByLeague(w http.ResponseWriter, r *http.Request) {
	leagueID, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.svc.ListByLeague(r.Context(), leagueID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /entries/{id}/roster  -> one entry with its players
func (h *RosterHandler) GetByEntry(w http.ResponseWriter, r *http.Request) {
	entryID, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.svc.GetByEntry(r.Context(), entryID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
