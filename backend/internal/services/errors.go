package services

import "errors"

var (
	// ErrNotFound is returned when a requested resource does not exist or is inaccessible.
	ErrNotFound = errors.New("not found")
	// ErrUnauthorized is returned when an operation is not permitted for the caller.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidID is returned when a provided ID cannot be parsed.
	ErrInvalidID = errors.New("invalid id")
	// ErrSessionExpired is returned when the session token has passed its expiry time.
	ErrSessionExpired = errors.New("session expired")
	// ErrCategoryInUse is returned when a category cannot be deleted because transactions reference it.
	ErrCategoryInUse = errors.New("category in use")
)
