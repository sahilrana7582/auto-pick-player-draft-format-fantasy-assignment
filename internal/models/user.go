package models

import "time"

// User is the DB model: it maps 1:1 to a row in the `users` table.
// Only the repository layer should create/read this directly.
type User struct {
	ID        int64
	FirstName string
	LastName  string
	Age       int
	CreatedAt time.Time
}
