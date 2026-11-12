package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Upper layers check for these instead of pgx-specific errors.
var (
	ErrNotFound = errors.New("not found") // row doesn't exist
	ErrConflict = errors.New("conflict")  // UNIQUE constraint violated

	// League rules (checked inside transactions, under the league row lock).
	ErrLeagueNotOpen      = errors.New("league is not in 'created' status")
	ErrLeagueFull         = errors.New("league is full")
	ErrLeagueNotFull      = errors.New("league does not have enough entries")
	ErrLeagueSizeTooSmall = errors.New("size is smaller than current number of entries")
)

// isUniqueViolation reports whether err is a Postgres UNIQUE violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
