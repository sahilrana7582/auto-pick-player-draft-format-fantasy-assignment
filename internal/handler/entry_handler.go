package handler

import (
	"encoding/json"
	"net/http"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/service"
)

// EntryHandler only deals with HTTP: parse input, call service, write output.
type EntryHandler struct {
	svc *service.EntryService
}

func NewEntryHandler(svc *service.EntryService) *EntryHandler {
	return &EntryHandler{svc: svc}
}

// RegisterRoutes wires the entry endpoints onto the mux.
// Joining/listing is nested under the league; single-entry ops use /entries/{id}.
func (h *EntryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /leagues/{id}/entries", h.Join)
	mux.HandleFunc("GET /leagues/{id}/entries", h.ListByLeague)
	mux.HandleFunc("GET /users/{id}/entries", h.ListByUser)
	mux.HandleFunc("GET /entries/{id}", h.Get)
	mux.HandleFunc("PUT /entries/{id}", h.Update)
	mux.HandleFunc("DELETE /entries/{id}", h.Leave)
}

// POST /leagues/{id}/entries  body: {"user_id": 1}
func (h *EntryHandler) Join(w http.ResponseWriter, r *http.Request) {
	leagueID, ok := parseID(w, r)
	if !ok {
		return
	}
	var req dto.CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.svc.Join(r.Context(), leagueID, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GET /leagues/{id}/entries  -> entries in draft order
func (h *EntryHandler) ListByLeague(w http.ResponseWriter, r *http.Request) {
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

// GET /users/{id}/entries  -> leagues this user has joined
func (h *EntryHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /entries/{id}
func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.svc.Get(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// PUT /entries/{id}  body: {"pick_order": 2}
func (h *EntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req dto.UpdateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// DELETE /entries/{id}  -> user leaves the league
func (h *EntryHandler) Leave(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Leave(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
