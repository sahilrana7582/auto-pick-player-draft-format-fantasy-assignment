package service

import (
	"context"
	"errors"
	"fmt"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// RosterService builds each entry's team for a league.
type RosterService struct {
	rosters *repository.RosterRepository
	entries *repository.EntryRepository
	leagues *repository.LeagueRepository
}

func NewRosterService(
	rosters *repository.RosterRepository,
	entries *repository.EntryRepository,
	leagues *repository.LeagueRepository,
) *RosterService {
	return &RosterService{rosters: rosters, entries: entries, leagues: leagues}
}

// ListByLeague returns every entry in draft order, each with the players it owns.
func (s *RosterService) ListByLeague(ctx context.Context, leagueID int64) ([]dto.EntryRosterResponse, error) {
	if _, err := s.leagues.GetByID(ctx, leagueID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("league %w", ErrNotFound)
		}
		return nil, err
	}

	// 1. All entries (so teams with no picks yet still show up).
	entries, err := s.entries.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	// 2. All picked players in the league.
	rosters, err := s.rosters.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	// 3. Group players by entry.
	playersByEntry := make(map[int64][]dto.RosterPlayerResponse)
	for _, ro := range rosters {
		playersByEntry[ro.EntryID] = append(playersByEntry[ro.EntryID], dto.ToRosterPlayerResponse(ro))
	}

	resp := make([]dto.EntryRosterResponse, 0, len(entries))
	for _, e := range entries {
		players := playersByEntry[e.ID]
		if players == nil {
			players = []dto.RosterPlayerResponse{} // JSON [] instead of null
		}
		resp = append(resp, dto.EntryRosterResponse{
			EntryID:   e.ID,
			UserID:    e.UserID,
			PickOrder: e.PickOrder,
			Players:   players,
		})
	}
	return resp, nil
}

// GetByEntry returns one entry with the players it owns.
// (No ownership check: any caller can view any entry's roster.)
func (s *RosterService) GetByEntry(ctx context.Context, entryID int64) (dto.EntryRosterResponse, error) {
	e, err := s.entries.GetByID(ctx, entryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.EntryRosterResponse{}, fmt.Errorf("entry %w", ErrNotFound)
		}
		return dto.EntryRosterResponse{}, err
	}

	rosters, err := s.rosters.ListByEntry(ctx, entryID)
	if err != nil {
		return dto.EntryRosterResponse{}, err
	}

	players := make([]dto.RosterPlayerResponse, 0, len(rosters)) // [] not null when empty
	for _, ro := range rosters {
		players = append(players, dto.ToRosterPlayerResponse(ro))
	}

	return dto.EntryRosterResponse{
		EntryID:   e.ID,
		UserID:    e.UserID,
		PickOrder: e.PickOrder,
		Players:   players,
	}, nil
}
