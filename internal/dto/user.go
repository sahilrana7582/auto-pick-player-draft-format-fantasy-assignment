package dto

import (
	"time"

	"auto-pick-player-draft-format-fantasy-assignment/internal/models"
)

// CreateUserRequest is the JSON body for POST /users.
type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}

// UpdateUserRequest is the JSON body for PUT /users/{id}.
// Kept the same shape as create: PUT replaces the whole user.
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}

// UserResponse is what we send back to the client.
// Separate from the DB model so DB changes don't leak into the API.
type UserResponse struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// ToUserResponse converts a DB model into an API response.
func ToUserResponse(u models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Age:       u.Age,
		CreatedAt: u.CreatedAt,
	}
}
