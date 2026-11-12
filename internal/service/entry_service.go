package service

import (
	"context"
	"errors"
	"fmt"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// EntryService handles users joining/leaving leagues.
// It uses the user & league repos only to give clear "X not found" errors.
type EntryService struct {
	entries *repository.EntryRepository
	users   *repository.UserRepository
	leagues *repository.LeagueRepository
}

func NewEntryService(
	entries *repository.EntryRepository,
	users *repository.UserRepository,
	leagues *repository.LeagueRepository,
) *EntryService {
	return &EntryService{entries: entries, users: users, leagues: leagues}
}

// Join adds a user to a league. pick_order is assigned automatically.
func (s *EntryService) Join(ctx context.Context, leagueID int64, req dto.CreateEntryRequest) (dto.EntryResponse, error) {
	if req.UserID <= 0 {
		return dto.EntryResponse{}, fmt.Errorf("%w: user_id is required", ErrValidation)
	}
	if err := s.ensureLeagueExists(ctx, leagueID); err != nil {
		return dto.EntryResponse{}, err
	}
	if err := s.ensureUserExists(ctx, req.UserID); err != nil {
		return dto.EntryResponse{}, err
	}

	e := models.Entry{LeagueID: leagueID, UserID: req.UserID}
	if err := s.entries.Create(ctx, &e); err != nil {
		switch {
		case errors.Is(err, repository.ErrConflict):
			return dto.EntryResponse{}, fmt.Errorf("%w: user has already joined this league", ErrConflict)
		case errors.Is(err, repository.ErrNotFound): // league deleted between check and insert
			return dto.EntryResponse{}, fmt.Errorf("league %w", ErrNotFound)
		}
		return dto.EntryResponse{}, err
	}
	return dto.ToEntryResponse(e), nil
}

func (s *EntryService) Get(ctx context.Context, id int64) (dto.EntryResponse, error) {
	e, err := s.entries.GetByID(ctx, id)
	if err != nil {
		return dto.EntryResponse{}, entryErr(err)
	}
	return dto.ToEntryResponse(e), nil
}

// ListByLeague returns the league's entries in draft order.
func (s *EntryService) ListByLeague(ctx context.Context, leagueID int64) ([]dto.EntryResponse, error) {
	if err := s.ensureLeagueExists(ctx, leagueID); err != nil {
		return nil, err
	}
	entries, err := s.entries.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	return dto.ToEntryResponses(entries), nil
}

// ListByUser returns every league entry for a user.
func (s *EntryService) ListByUser(ctx context.Context, userID int64) ([]dto.EntryResponse, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}
	entries, err := s.entries.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.ToEntryResponses(entries), nil
}

// Update changes the entry's pick_order (reorder the draft).
func (s *EntryService) Update(ctx context.Context, id int64, req dto.UpdateEntryRequest) (dto.EntryResponse, error) {
	if req.PickOrder <= 0 {
		return dto.EntryResponse{}, fmt.Errorf("%w: pick_order must be greater than 0", ErrValidation)
	}
	e, err := s.entries.UpdatePickOrder(ctx, id, req.PickOrder)
	if errors.Is(err, repository.ErrConflict) {
		return dto.EntryResponse{}, fmt.Errorf("%w: pick_order %d is already taken in this league", ErrConflict, req.PickOrder)
	}
	if err != nil {
		return dto.EntryResponse{}, entryErr(err)
	}
	return dto.ToEntryResponse(e), nil
}

// Leave removes the entry. Note: this leaves a gap in pick_order (e.g. 1, 3, 4);
// the draft should always iterate entries ORDER BY pick_order, so gaps are fine.
func (s *EntryService) Leave(ctx context.Context, id int64) error {
	return entryErr(s.entries.Delete(ctx, id))
}

func (s *EntryService) ensureLeagueExists(ctx context.Context, id int64) error {
	if _, err := s.leagues.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("league %w", ErrNotFound)
		}
		return err
	}
	return nil
}

func (s *EntryService) ensureUserExists(ctx context.Context, id int64) error {
	if _, err := s.users.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("user %w", ErrNotFound)
		}
		return err
	}
	return nil
}

// entryErr turns a repo not-found into "entry not found".
func entryErr(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("entry %w", ErrNotFound)
	}
	return mapRepoErr(err)
}
