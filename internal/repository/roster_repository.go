package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// RosterRepository reads roster_owners.
type RosterRepository struct {
	db *pgxpool.Pool
}

func NewRosterRepository(db *pgxpool.Pool) *RosterRepository {
	return &RosterRepository{db: db}
}

// Shared SELECT: roster rows joined with player details.
// Callers add the WHERE; results are in the order players were picked.
const rosterSelect = `
	SELECT ro.id, ro.league_id, ro.entry_id, ro.player_id, ro.created_at,
	       p.id, p.name, p.position, p.created_at
	FROM roster_owners ro
	JOIN players p ON p.id = ro.player_id`

const rosterOrder = ` ORDER BY ro.created_at, ro.id`

// ListByLeague returns every picked player in the league.
func (r *RosterRepository) ListByLeague(ctx context.Context, leagueID int64) ([]models.RosterOwner, error) {
	return r.list(ctx, rosterSelect+` WHERE ro.league_id = $1`+rosterOrder, leagueID)
}

// ListByEntry returns the players owned by one entry.
func (r *RosterRepository) ListByEntry(ctx context.Context, entryID int64) ([]models.RosterOwner, error) {
	return r.list(ctx, rosterSelect+` WHERE ro.entry_id = $1`+rosterOrder, entryID)
}

func (r *RosterRepository) list(ctx context.Context, q string, arg int64) ([]models.RosterOwner, error) {
	rows, err := r.db.Query(ctx, q, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rosters := []models.RosterOwner{}
	for rows.Next() {
		var ro models.RosterOwner
		if err := rows.Scan(
			&ro.ID, &ro.LeagueID, &ro.EntryID, &ro.PlayerID, &ro.CreatedAt,
			&ro.Player.ID, &ro.Player.Name, &ro.Player.Position, &ro.Player.CreatedAt,
		); err != nil {
			return nil, err
		}
		rosters = append(rosters, ro)
	}
	return rosters, rows.Err()
}
