package domain

import "errors"

var (
	ErrInvalidTransition    = errors.New("invalid status transition")
	ErrForbidden            = errors.New("action not permitted for this user")
	ErrTargetRequired       = errors.New("target_word_id is required for this contribution type")
	ErrContributionNotFound = errors.New("contribution not found")
	ErrVoteNotFound         = errors.New("vote not found")
	ErrBookmarkNotFound     = errors.New("bookmark not found")
	ErrBookmarkConflict     = errors.New("bookmark already exists")
	ErrCommentNotFound      = errors.New("comment not found")
	ErrBodyTooLong          = errors.New("comment body exceeds 1000 characters")
	ErrEmptyBody            = errors.New("comment body must not be empty")
	ErrInvalidVoteTarget    = errors.New("invalid vote target type; must be 'word' or 'definition'")
	ErrInvalidVoteValue     = errors.New("invalid vote value; must be 'up' or 'down'")
)
