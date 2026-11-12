package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// EntryRepository only talks to the DB. No business rules here.
type EntryRepository struct {
	db *pgxpool.Pool
}

func NewEntryRepository(db *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{db: db}
}

const entryColumns = `id, league_id, user_id, pick_order, joined_at, created_at`

func scanEntry(row pgx.Row) (models.Entry, error) {
	var e models.Entry
	err := row.Scan(&e.ID, &e.LeagueID, &e.UserID, &e.PickOrder, &e.JoinedAt, &e.CreatedAt)
	return e, err
}

// Create adds the user to the league with the next pick_order (max + 1).
//
// Why a transaction + lock?
// Two users joining the same league at the same time could both read
// MAX(pick_order) = 3 and both try to insert 4. Locking the league row
// (SELECT ... FOR UPDATE) makes concurrent joins for the SAME league wait
// in line, so each one sees the previous insert. Joins to other leagues
// are not blocked.
func (r *EntryRepository) Create(ctx context.Context, e *models.Entry) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op after a successful Commit

	// 1. Lock the league row (also confirms the league exists) and read status + size.
	var status models.LeagueStatus
	var size int
	err = tx.QueryRow(ctx, `SELECT status, size FROM leagues WHERE id = $1 FOR UPDATE`, e.LeagueID).
		Scan(&status, &size)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// These checks must happen here, under the lock: "start league" and other
	// joins lock the same row, so nothing can slip in between check and insert.

	// Only 'created' leagues accept new entries.
	if status != models.LeagueStatusCreated {
		return ErrLeagueNotOpen
	}

	// League can't have more entries than its size.
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE league_id = $1`, e.LeagueID).Scan(&count); err != nil {
		return err
	}
	if count >= size {
		return ErrLeagueFull
	}

	// 2. Insert with the next pick_order for this league.
	const q = `
		INSERT INTO entries (league_id, user_id, pick_order)
		SELECT $1, $2, COALESCE(MAX(pick_order), 0) + 1
		FROM entries
		WHERE league_id = $1
		RETURNING ` + entryColumns
	created, err := scanEntry(tx.QueryRow(ctx, q, e.LeagueID, e.UserID))
	if isUniqueViolation(err) {
		return ErrConflict // user already joined this league
	}
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	*e = created
	return nil
}

func (r *EntryRepository) GetByID(ctx context.Context, id int64) (models.Entry, error) {
	e, err := scanEntry(r.db.QueryRow(ctx, `SELECT `+entryColumns+` FROM entries WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Entry{}, ErrNotFound
	}
	return e, err
}

// ListByLeague returns all entries in a league in draft order.
func (r *EntryRepository) ListByLeague(ctx context.Context, leagueID int64) ([]models.Entry, error) {
	return r.list(ctx, `SELECT `+entryColumns+` FROM entries WHERE league_id = $1 ORDER BY pick_order`, leagueID)
}

// ListByUser returns all leagues a user has joined (newest first).
func (r *EntryRepository) ListByUser(ctx context.Context, userID int64) ([]models.Entry, error) {
	return r.list(ctx, `SELECT `+entryColumns+` FROM entries WHERE user_id = $1 ORDER BY joined_at DESC`, userID)
}

func (r *EntryRepository) list(ctx context.Context, q string, arg int64) ([]models.Entry, error) {
	rows, err := r.db.Query(ctx, q, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.Entry{}
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// UpdatePickOrder changes an entry's turn order.
// Returns ErrConflict if another entry in the league already has that order.
func (r *EntryRepository) UpdatePickOrder(ctx context.Context, id int64, pickOrder int) (models.Entry, error) {
	const q = `UPDATE entries SET pick_order = $1 WHERE id = $2 RETURNING ` + entryColumns
	e, err := scanEntry(r.db.QueryRow(ctx, q, pickOrder, id))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Entry{}, ErrNotFound
	case isUniqueViolation(err):
		return models.Entry{}, ErrConflict
	}
	return e, err
}

// Delete removes the entry (user leaves the league). Their roster rows cascade.
func (r *EntryRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
