package service

import (
	"context"
	"fmt"
	"strings"

	"auto-pick-player-draft-format-fantasy-assignment/internal/dto"
	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
)

// UserService holds business logic: validation + converting request -> model -> response.
// It knows nothing about HTTP.
type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest) (dto.UserResponse, error) {
	u := models.User{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Age:       req.Age,
	}
	if err := validateUser(u); err != nil {
		return dto.UserResponse{}, err
	}
	if err := s.repo.Create(ctx, &u); err != nil {
		return dto.UserResponse{}, err
	}
	return dto.ToUserResponse(u), nil
}

func (s *UserService) Get(ctx context.Context, id int64) (dto.UserResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.UserResponse{}, mapRepoErr(err)
	}
	return dto.ToUserResponse(u), nil
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]dto.UserResponse, error) {
	limit, offset = normalizePage(limit, offset)

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, dto.ToUserResponse(u))
	}
	return resp, nil
}

func (s *UserService) Update(ctx context.Context, id int64, req dto.UpdateUserRequest) (dto.UserResponse, error) {
	u := models.User{
		ID:        id,
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Age:       req.Age,
	}
	if err := validateUser(u); err != nil {
		return dto.UserResponse{}, err
	}
	if err := s.repo.Update(ctx, &u); err != nil {
		return dto.UserResponse{}, mapRepoErr(err)
	}
	return dto.ToUserResponse(u), nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return mapRepoErr(s.repo.Delete(ctx, id))
}

// validateUser checks the same rules the DB enforces, so we can give a clear 400.
func validateUser(u models.User) error {
	switch {
	case u.FirstName == "":
		return fmt.Errorf("%w: first_name is required", ErrValidation)
	case u.LastName == "":
		return fmt.Errorf("%w: last_name is required", ErrValidation)
	case u.Age <= 0:
		return fmt.Errorf("%w: age must be greater than 0", ErrValidation)
	}
	return nil
}
