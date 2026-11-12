package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// UserRepository only talks to the DB. No business rules here.
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a user and fills in the generated ID and created_at.
func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	const q = `
		INSERT INTO users (first_name, last_name, age)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(ctx, q, u.FirstName, u.LastName, u.Age).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (models.User, error) {
	const q = `
		SELECT id, first_name, last_name, age, created_at
		FROM users
		WHERE id = $1`
	var u models.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Age, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

// List returns users ordered by id (simple limit/offset pagination).
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	const q = `
		SELECT id, first_name, last_name, age, created_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Age, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Update overwrites name/age. Returns ErrNotFound if the id doesn't exist.
func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	const q = `
		UPDATE users
		SET first_name = $1, last_name = $2, age = $3
		WHERE id = $4
		RETURNING created_at`
	err := r.db.QueryRow(ctx, q, u.FirstName, u.LastName, u.Age, u.ID).Scan(&u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Delete removes a user. Returns ErrNotFound if nothing was deleted.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
