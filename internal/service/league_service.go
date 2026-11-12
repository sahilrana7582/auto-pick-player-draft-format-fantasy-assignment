package service

import (
	"context"
	"fmt"
	"strings"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// LeagueService holds league business logic (validation, mapping). No HTTP here.
type LeagueService struct {
	repo *repository.LeagueRepository
}

func NewLeagueService(repo *repository.LeagueRepository) *LeagueService {
	return &LeagueService{repo: repo}
}

func (s *LeagueService) Create(ctx context.Context, req dto.CreateLeagueRequest) (dto.LeagueResponse, error) {
	l := buildLeague(req.Name, req.Description, req.Type, req.SeriesName)
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
	l := buildLeague(req.Name, req.Description, req.Type, req.SeriesName)
	l.ID = id
	if err := validateLeague(l); err != nil {
		return dto.LeagueResponse{}, err
	}
	if err := s.repo.Update(ctx, &l); err != nil {
		return dto.LeagueResponse{}, mapRepoErr(err)
	}
	return dto.ToLeagueResponse(l), nil
}

func (s *LeagueService) Delete(ctx context.Context, id int64) error {
	return mapRepoErr(s.repo.Delete(ctx, id))
}

// buildLeague trims input and turns an empty description into NULL.
func buildLeague(name, description, leagueType, seriesName string) models.League {
	l := models.League{
		Name:       strings.TrimSpace(name),
		Type:       models.LeagueType(strings.TrimSpace(leagueType)),
		SeriesName: strings.TrimSpace(seriesName),
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
	}
	return nil
}
