package repository

import "errors"

// ErrNotFound is returned when a row doesn't exist.
// Upper layers check for this instead of pgx-specific errors.
var ErrNotFound = errors.New("not found")
