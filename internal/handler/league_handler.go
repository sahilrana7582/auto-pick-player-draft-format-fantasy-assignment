package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/service"
)

// LeagueHandler only deals with HTTP: parse input, call service, write output.
type LeagueHandler struct {
	svc *service.LeagueService
}

func NewLeagueHandler(svc *service.LeagueService) *LeagueHandler {
	return &LeagueHandler{svc: svc}
}

// RegisterRoutes wires the league endpoints onto the mux.
func (h *LeagueHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /leagues", h.Create)
	mux.HandleFunc("GET /leagues", h.List)
	mux.HandleFunc("GET /leagues/{id}", h.Get)
	mux.HandleFunc("PUT /leagues/{id}", h.Update)
	mux.HandleFunc("DELETE /leagues/{id}", h.Delete)
}

// POST /leagues
func (h *LeagueHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLeagueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.svc.Create(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GET /leagues?limit=20&offset=0
func (h *LeagueHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	resp, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /leagues/{id}
func (h *LeagueHandler) Get(w http.ResponseWriter, r *http.Request) {
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

// PUT /leagues/{id}
func (h *LeagueHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req dto.UpdateLeagueRequest
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

// DELETE /leagues/{id}
func (h *LeagueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
