package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// LeagueRepository talks to the DB for leagues.
type LeagueRepository struct {
	db *pgxpool.Pool
}

func NewLeagueRepository(db *pgxpool.Pool) *LeagueRepository {
	return &LeagueRepository{db: db}
}

const leagueColumns = `id, name, description, type, series_name, status, size, timeout, created_at`

func scanLeague(row pgx.Row) (models.League, error) {
	var l models.League
	err := row.Scan(&l.ID, &l.Name, &l.Description, &l.Type, &l.SeriesName,
		&l.Status, &l.Size, &l.Timeout, &l.CreatedAt)
	return l, err
}

// Create inserts a league. status defaults to 'created' in the DB.
func (r *LeagueRepository) Create(ctx context.Context, l *models.League) error {
	const q = `
		INSERT INTO leagues (name, description, type, series_name, size, timeout)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + leagueColumns
	created, err := scanLeague(r.db.QueryRow(ctx, q,
		l.Name, l.Description, l.Type, l.SeriesName, l.Size, l.Timeout))
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

// Update overwrites the editable fields (not status).
//
// The WHERE also guards size: you can't shrink a league below the number
// of entries that already joined (e.g. 5 joined -> size can't become 4).
// Returns ErrNotFound if the id doesn't exist, ErrLeagueSizeTooSmall for the size guard.
func (r *LeagueRepository) Update(ctx context.Context, l *models.League) error {
	const q = `
		UPDATE leagues
		SET name = $1, description = $2, type = $3, series_name = $4, size = $5, timeout = $6
		WHERE id = $7
		  AND $5 >= (SELECT COUNT(*) FROM entries WHERE league_id = $7)
		RETURNING ` + leagueColumns
	updated, err := scanLeague(r.db.QueryRow(ctx, q,
		l.Name, l.Description, l.Type, l.SeriesName, l.Size, l.Timeout, l.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		// Nothing updated: either the league doesn't exist or the size guard failed.
		if _, getErr := r.GetByID(ctx, l.ID); getErr != nil {
			return getErr
		}
		return ErrLeagueSizeTooSmall
	}
	if err != nil {
		return err
	}
	*l = updated
	return nil
}

// Start begins the draft, all in ONE transaction:
//
//  1. lock the league row           (joins/starts for this league wait for us)
//  2. status must be 'created'      -> else ErrLeagueNotOpen
//  3. entries count must == size    -> else ErrLeagueNotFull
//  4. first entry = ORDER BY pick_order LIMIT 1
//  5. status -> 'started'
//  6. create the league_picks row: first entry's turn, expires in `timeout` seconds
//
// If anything fails, the whole thing rolls back (no half-started league).
func (r *LeagueRepository) Start(ctx context.Context, leagueID int64) (models.League, models.LeaguePick, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.League{}, models.LeaguePick{}, err
	}
	defer tx.Rollback(ctx) // no-op after a successful Commit

	// 1. Lock + read the league.
	var status models.LeagueStatus
	var size int
	err = tx.QueryRow(ctx, `SELECT status, size FROM leagues WHERE id = $1 FOR UPDATE`, leagueID).
		Scan(&status, &size)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.League{}, models.LeaguePick{}, ErrNotFound
	}
	if err != nil {
		return models.League{}, models.LeaguePick{}, err
	}

	// 2. Only a 'created' league can start.
	if status != models.LeagueStatusCreated {
		return models.League{}, models.LeaguePick{}, ErrLeagueNotOpen
	}

	// 3. League must be full.
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE league_id = $1`, leagueID).Scan(&count); err != nil {
		return models.League{}, models.LeaguePick{}, err
	}
	if count < size {
		return models.League{}, models.LeaguePick{}, fmt.Errorf("%w: %d of %d joined", ErrLeagueNotFull, count, size)
	}

	// 4. The entry with the lowest pick_order picks first.
	var firstEntryID int64
	err = tx.QueryRow(ctx,
		`SELECT id FROM entries WHERE league_id = $1 ORDER BY pick_order LIMIT 1`, leagueID).
		Scan(&firstEntryID)
	if err != nil {
		return models.League{}, models.LeaguePick{}, err
	}

	// 5. Mark the league as started.
	league, err := scanLeague(tx.QueryRow(ctx,
		`UPDATE leagues SET status = 'started' WHERE id = $1 RETURNING `+leagueColumns, leagueID))
	if err != nil {
		return models.League{}, models.LeaguePick{}, err
	}

	// 6. First turn: deadline = now + timeout seconds.
	var pick models.LeaguePick
	err = tx.QueryRow(ctx, `
		INSERT INTO league_picks (league_id, entry_id, expires_at)
		VALUES ($1, $2, NOW() + make_interval(secs => $3))
		RETURNING id, league_id, entry_id, expires_at, updated_at`,
		leagueID, firstEntryID, league.Timeout,
	).Scan(&pick.ID, &pick.LeagueID, &pick.EntryID, &pick.ExpiresAt, &pick.UpdatedAt)
	if err != nil {
		return models.League{}, models.LeaguePick{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.League{}, models.LeaguePick{}, err
	}
	return league, pick, nil
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
