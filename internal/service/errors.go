package service

import (
	"errors"

	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// Errors the handler can map to HTTP status codes.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

// mapRepoErr converts repository errors into service errors.
func mapRepoErr(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
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
