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

// Create inserts a league and fills in the generated ID and created_at.
func (r *LeagueRepository) Create(ctx context.Context, l *models.League) error {
	const q = `
		INSERT INTO leagues (name, description, type, series_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return r.db.QueryRow(ctx, q, l.Name, l.Description, l.Type, l.SeriesName).
		Scan(&l.ID, &l.CreatedAt)
}

func (r *LeagueRepository) GetByID(ctx context.Context, id int64) (models.League, error) {
	const q = `
		SELECT id, name, description, type, series_name, created_at
		FROM leagues
		WHERE id = $1`
	var l models.League
	err := r.db.QueryRow(ctx, q, id).
		Scan(&l.ID, &l.Name, &l.Description, &l.Type, &l.SeriesName, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.League{}, ErrNotFound
	}
	return l, err
}

// List returns leagues ordered by id (simple limit/offset pagination).
func (r *LeagueRepository) List(ctx context.Context, limit, offset int) ([]models.League, error) {
	const q = `
		SELECT id, name, description, type, series_name, created_at
		FROM leagues
		ORDER BY id
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leagues := []models.League{}
	for rows.Next() {
		var l models.League
		if err := rows.Scan(&l.ID, &l.Name, &l.Description, &l.Type, &l.SeriesName, &l.CreatedAt); err != nil {
			return nil, err
		}
		leagues = append(leagues, l)
	}
	return leagues, rows.Err()
}

// Update overwrites all editable fields. Returns ErrNotFound if the id doesn't exist.
func (r *LeagueRepository) Update(ctx context.Context, l *models.League) error {
	const q = `
		UPDATE leagues
		SET name = $1, description = $2, type = $3, series_name = $4
		WHERE id = $5
		RETURNING created_at`
	err := r.db.QueryRow(ctx, q, l.Name, l.Description, l.Type, l.SeriesName, l.ID).
		Scan(&l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
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
