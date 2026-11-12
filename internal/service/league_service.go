package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// Same limits as the CHECK constraints in the leagues migration.
const (
	MinLeagueSize  = 2   // a draft needs at least 2 entries
	MaxPickTimeout = 300 // seconds
)

// LeagueService holds league business logic (validation, mapping). No HTTP here.
type LeagueService struct {
	repo *repository.LeagueRepository
}

func NewLeagueService(repo *repository.LeagueRepository) *LeagueService {
	return &LeagueService{repo: repo}
}

func (s *LeagueService) Create(ctx context.Context, req dto.CreateLeagueRequest) (dto.LeagueResponse, error) {
	l := buildLeague(req.Name, req.Description, req.Type, req.SeriesName, req.Size, req.Timeout)
	if err := validateLeague(l); err != nil {
		return dto.LeagueResponse{}, err
	}
	if err := s.repo.Create(ctx, &l); err != nil {
		return dto.LeagueResponse{}, err
	}
	return dto.ToLeagueResponse(l), nil
}

func (s *LeagueService) Get(ctx context.Context, id int64) (dto.LeagueResponse, error) {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.LeagueResponse{}, mapRepoErr(err)
	}
	return dto.ToLeagueResponse(l), nil
}

func (s *LeagueService) List(ctx context.Context, limit, offset int) ([]dto.LeagueResponse, error) {
	limit, offset = normalizePage(limit, offset)

	leagues, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.LeagueResponse, 0, len(leagues))
	for _, l := range leagues {
		resp = append(resp, dto.ToLeagueResponse(l))
	}
	return resp, nil
}

func (s *LeagueService) Update(ctx context.Context, id int64, req dto.UpdateLeagueRequest) (dto.LeagueResponse, error) {
	l := buildLeague(req.Name, req.Description, req.Type, req.SeriesName, req.Size, req.Timeout)
	l.ID = id
	if err := validateLeague(l); err != nil {
		return dto.LeagueResponse{}, err
	}
	if err := s.repo.Update(ctx, &l); err != nil {
		if errors.Is(err, repository.ErrLeagueSizeTooSmall) {
			return dto.LeagueResponse{}, fmt.Errorf("%w: size is smaller than the number of entries already joined", ErrConflict)
		}
		return dto.LeagueResponse{}, mapRepoErr(err)
	}
	return dto.ToLeagueResponse(l), nil
}

// Start begins the draft: league must be 'created' and full (entries == size).
// On success the league is 'started' and the first entry (lowest pick_order)
// is on the clock for `timeout` seconds.
func (s *LeagueService) Start(ctx context.Context, id int64) (dto.StartLeagueResponse, error) {
	l, pick, err := s.repo.Start(ctx, id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return dto.StartLeagueResponse{}, fmt.Errorf("league %w", ErrNotFound)
	case errors.Is(err, repository.ErrLeagueNotOpen):
		return dto.StartLeagueResponse{}, fmt.Errorf("%w: league can only be started from 'created' status", ErrConflict)
	case errors.Is(err, repository.ErrLeagueNotFull):
		// err already says e.g. "league does not have enough entries: 1 of 4 joined"
		return dto.StartLeagueResponse{}, fmt.Errorf("%w: %v", ErrConflict, err)
	case err != nil:
		return dto.StartLeagueResponse{}, err
	}
	return dto.StartLeagueResponse{
		League:      dto.ToLeagueResponse(l),
		CurrentPick: dto.CurrentPickResponse{EntryID: pick.EntryID, ExpiresAt: pick.ExpiresAt},
	}, nil
}

func (s *LeagueService) Delete(ctx context.Context, id int64) error {
	return mapRepoErr(s.repo.Delete(ctx, id))
}

// buildLeague trims input and turns an empty description into NULL.
func buildLeague(name, description, leagueType, seriesName string, size, timeout int) models.League {
	l := models.League{
		Name:       strings.TrimSpace(name),
		Type:       models.LeagueType(strings.TrimSpace(leagueType)),
		SeriesName: strings.TrimSpace(seriesName),
		Size:       size,
		Timeout:    timeout,
	}
	if d := strings.TrimSpace(description); d != "" {
		l.Description = &d
	}
	return l
}

func validateLeague(l models.League) error {
	switch {
	case l.Name == "":
		return fmt.Errorf("%w: name is required", ErrValidation)
	case !l.Type.IsValid():
		return fmt.Errorf("%w: type must be 'season_long' or 'draft_pick'", ErrValidation)
	case l.SeriesName == "":
		return fmt.Errorf("%w: series_name is required", ErrValidation)
	case l.Size < MinLeagueSize:
		return fmt.Errorf("%w: size must be at least %d", ErrValidation, MinLeagueSize)
	case l.Timeout < 1 || l.Timeout > MaxPickTimeout:
		return fmt.Errorf("%w: timeout must be between 1 and %d seconds", ErrValidation, MaxPickTimeout)
	}
	return nil
}
