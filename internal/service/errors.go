package service

import "errors"

// Errors the handler can map to HTTP status codes.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)
