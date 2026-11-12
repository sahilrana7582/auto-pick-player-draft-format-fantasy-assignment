package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// LeagueRepository only talks to the DB. No business rules here.
type LeagueRepository struct {
	db *pgxpool.Pool
}

func NewLeagueRepository(db *pgxpool.Pool) *LeagueRepository {
	return &LeagueRepository{db: db}
}

const leagueColumns = `id, name, description, type, series_name, status, created_at`

func scanLeague(row pgx.Row) (models.League, error) {
	var l models.League
	err := row.Scan(&l.ID, &l.Name, &l.Description, &l.Type, &l.SeriesName, &l.Status, &l.CreatedAt)
	return l, err
}

// Create inserts a league. status defaults to 'created' in the DB.
func (r *LeagueRepository) Create(ctx context.Context, l *models.League) error {
	const q = `
		INSERT INTO leagues (name, description, type, series_name)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + leagueColumns
	created, err := scanLeague(r.db.QueryRow(ctx, q, l.Name, l.Description, l.Type, l.SeriesName))
	if err != nil {
		return err
	}
	*l = created
	return nil
}

func (r *LeagueRepository) GetByID(ctx context.Context, id int64) (models.League, error) {
	l, err := scanLeague(r.db.QueryRow(ctx, `SELECT `+leagueColumns+` FROM leagues WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.League{}, ErrNotFound
	}
	return l, err
}

// List returns leagues ordered by id (simple limit/offset pagination).
func (r *LeagueRepository) List(ctx context.Context, limit, offset int) ([]models.League, error) {
	const q = `SELECT ` + leagueColumns + ` FROM leagues ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leagues := []models.League{}
	for rows.Next() {
		l, err := scanLeague(rows)
		if err != nil {
			return nil, err
		}
		leagues = append(leagues, l)
	}
	return leagues, rows.Err()
}

// Update overwrites the editable fields (not status). Returns ErrNotFound if the id doesn't exist.
func (r *LeagueRepository) Update(ctx context.Context, l *models.League) error {
	const q = `
		UPDATE leagues
		SET name = $1, description = $2, type = $3, series_name = $4
		WHERE id = $5
		RETURNING ` + leagueColumns
	updated, err := scanLeague(r.db.QueryRow(ctx, q, l.Name, l.Description, l.Type, l.SeriesName, l.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	*l = updated
	return nil
}

// UpdateStatus moves a league from one status to another, atomically.
//
// The `AND status = from` check makes this a compare-and-swap: if two
// requests try to start the same league, only one UPDATE matches.
// Returns ErrNotFound if the league doesn't exist, ErrConflict if it's not in `from`.
func (r *LeagueRepository) UpdateStatus(ctx context.Context, id int64, from, to models.LeagueStatus) (models.League, error) {
	const q = `UPDATE leagues SET status = $1 WHERE id = $2 AND status = $3 RETURNING ` + leagueColumns
	l, err := scanLeague(r.db.QueryRow(ctx, q, to, id, from))
	if errors.Is(err, pgx.ErrNoRows) {
		// Nothing updated: either the league doesn't exist or it's in another status.
		if _, getErr := r.GetByID(ctx, id); getErr != nil {
			return models.League{}, getErr // ErrNotFound (or a real DB error)
		}
		return models.League{}, ErrConflict
	}
	return l, err
}

// Delete removes a league (entries/rosters/picks cascade via FK).
// Returns ErrNotFound if nothing was deleted.
func (r *LeagueRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM leagues WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
