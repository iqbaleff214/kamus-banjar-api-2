package commands

import "errors"

var (
	ErrWordNotFound         = errors.New("word not found")
	ErrContributionNotFound = errors.New("contribution not found")
	ErrAIRequestNotFound    = errors.New("ai request not found")
	ErrAIRequestConflict    = errors.New("ai request state conflict")
)
