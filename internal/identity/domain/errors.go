package domain

import "errors"

var (
	ErrEmptyName        = errors.New("name is required")
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidRole      = errors.New("invalid role")
	ErrUserBanned       = errors.New("user is banned")
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailConflict    = errors.New("email already registered")
	ErrWrongPassword    = errors.New("incorrect password")
	ErrUnverifiedEmail  = errors.New("email not verified")
)
