package workers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// leagueData is everything the worker needs to auto-pick inside one league.
type leagueData struct {
	picks            []models.LeaguePick // timed-out turns in this league
	availablePlayers []models.Player     // players NOT yet owned by any entry in this league
}

type AutoDraftPickWorker struct {
	pool      *pgxpool.Pool
	healthMap *sync.Map
}

func New(pool *pgxpool.Pool, parMap *sync.Map) *AutoDraftPickWorker {
	return &AutoDraftPickWorker{
		pool:      pool,
		healthMap: parMap,
	}
}

// Run checks for timed-out picks every `interval` until ctx is cancelled.
// It blocks, so start it in its own goroutine:
//
//	go worker.Run(ctx, 2*time.Second)
func (w *AutoDraftPickWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval) // wait between polls instead of hammering the DB
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.handleTimeOutPicks(ctx); err != nil {
				log.Printf("auto-pick worker: %v", err) // log and keep going; next tick retries
			}
		}
	}
}

// handleTimeOutPicks finds every turn whose deadline has passed, groups them
// by league, loads that league's available players, and hands off to handleSingleLeague.
func (w *AutoDraftPickWorker) handleTimeOutPicks(ctx context.Context) error {
	timeoutPicks, err := w.fetchTimedOutPicks(ctx)
	if err != nil {
		return err
	}
	if len(timeoutPicks) == 0 {
		return nil
	}

	// Group the timed-out picks by league.
	// Pointer values so we can append in place without re-storing into the map.
	leagueEntriesMap := make(map[int64]*leagueData)
	for _, pick := range timeoutPicks {
		lData, ok := leagueEntriesMap[pick.LeagueID]
		if !ok {
			lData = &leagueData{}
			leagueEntriesMap[pick.LeagueID] = lData
		}
		lData.picks = append(lData.picks, pick)
	}

	for leagueID, lData := range leagueEntriesMap {
		players, err := w.fetchAvailablePlayers(ctx, leagueID)
		if err != nil {
			// One broken league shouldn't stop the others from being processed.
			log.Printf("auto-pick worker: league %d: load available players: %v", leagueID, err)
			continue
		}
		lData.availablePlayers = players

		if err := w.handleSingleLeague(ctx, leagueID, lData); err != nil {
			log.Printf("auto-pick worker: league %d: %v", leagueID, err)
		}
	}

	return nil
}

// fetchTimedOutPicks returns turns whose deadline has passed (oldest first).
func (w *AutoDraftPickWorker) fetchTimedOutPicks(ctx context.Context) ([]models.LeaguePick, error) {
	// Explicit columns (not SELECT *) so the Scan order can't silently break.
	const query = `
		SELECT id, league_id, entry_id, expires_at, updated_at
		FROM league_picks
		WHERE expires_at <= NOW()
		ORDER BY expires_at`

	rows, err := w.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeoutPicks []models.LeaguePick
	for rows.Next() {
		var p models.LeaguePick
		if err := rows.Scan(&p.ID, &p.LeagueID, &p.EntryID, &p.ExpiresAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		timeoutPicks = append(timeoutPicks, p)
	}
	return timeoutPicks, rows.Err()
}

// fetchAvailablePlayers returns players nobody in this league has picked yet.
// Ownership lives in roster_owners (league_picks only tracks whose turn it is).
func (w *AutoDraftPickWorker) fetchAvailablePlayers(ctx context.Context, leagueID int64) ([]models.Player, error) {
	const query = `
		SELECT p.id, p.name, p.position, p.created_at
		FROM players p
		WHERE NOT EXISTS (
			SELECT 1
			FROM roster_owners ro
			WHERE ro.player_id = p.id
			  AND ro.league_id = $1
		)
		ORDER BY p.id`

	rows, err := w.pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // in its own function so it closes per league, not at the end of the loop

	var players []models.Player
	for rows.Next() {
		var p models.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.Position, &p.CreatedAt); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// handleSingleLeague auto-picks for every timed-out turn in one league.
// Strategy: always give the first available player (lowest id).
func (w *AutoDraftPickWorker) handleSingleLeague(ctx context.Context, leagueID int64, lData *leagueData) error {
	availablePlayers := lData.availablePlayers

	// Name the columns: roster_owners also has `id` (auto) and `created_at` (default),
	// so without a column list Postgres would try to put $1 into `id`.
	const query = `
		INSERT INTO roster_owners (league_id, entry_id, player_id)
		VALUES ($1, $2, $3)`

	for _, pick := range lData.picks {
		// Nobody left to pick.
		if len(availablePlayers) == 0 {
			return nil
		}

		// First available player.
		player := availablePlayers[0]

		if _, err := w.pool.Exec(ctx, query, pick.LeagueID, pick.EntryID, player.ID); err != nil {
			return fmt.Errorf("entry %d pick player %d: %w", pick.EntryID, player.ID, err)
		}

		// That player is taken now; the next pick gets the next one.
		availablePlayers = availablePlayers[1:]
	}
	return nil
}
