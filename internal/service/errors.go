package service

import (
	"errors"

	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// Errors the handler can map to HTTP status codes.
// Wrap them with context, e.g. fmt.Errorf("league %w", ErrNotFound) -> "league not found".
var (
	ErrNotFound   = errors.New("not found")        // 404
	ErrValidation = errors.New("validation error") // 400
	ErrConflict   = errors.New("conflict")         // 409
)

// mapRepoErr converts repository errors into service errors.
func mapRepoErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrConflict):
		return ErrConflict
	}
	return err
}

// normalizePage keeps pagination sane: default 20, max 100, no negative offset.
func normalizePage(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
