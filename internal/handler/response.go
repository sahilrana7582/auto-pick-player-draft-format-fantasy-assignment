package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"auto-pick-player-draft-format-fantasy-assignment/internal/service"
)

// writeJSON sends any value as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError sends {"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseID reads {id} from the path; writes a 400 and returns false if invalid.
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

// handleServiceError maps service errors to HTTP status codes.
// Unknown errors are logged and hidden behind a generic 500.
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error()) // e.g. "league not found"
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
