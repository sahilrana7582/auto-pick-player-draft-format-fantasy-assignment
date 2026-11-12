package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Upper layers check for these instead of pgx-specific errors.
var (
	ErrNotFound      = errors.New("not found")                      // row doesn't exist
	ErrConflict      = errors.New("conflict")                       // UNIQUE constraint violated / wrong status
	ErrLeagueNotOpen = errors.New("league is not open for joining") // league status != 'created'
)

// isUniqueViolation reports whether err is a Postgres UNIQUE violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
