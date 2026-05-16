package commands

import "errors"

var (
	ErrContributionNotFound = errors.New("contribution not found")
	ErrContributionConflict = errors.New("contribution is not pending")
	ErrNoteRequired         = errors.New("reviewer note is required")
	ErrUserNotFound         = errors.New("user not found")
	ErrWordNotFound         = errors.New("word not found")
	ErrInvalidRole          = errors.New("invalid role value; must be 'user' or 'admin'")
	ErrInvalidPayload       = errors.New("contribution payload is malformed")
)
