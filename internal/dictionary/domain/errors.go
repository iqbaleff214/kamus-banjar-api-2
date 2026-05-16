package domain

import "errors"

var (
	ErrEmptyBanjar      = errors.New("banjar word is required")
	ErrInvalidWordClass = errors.New("invalid word class")
	ErrInvalidDialect   = errors.New("invalid dialect")
	ErrInvalidSource    = errors.New("invalid source")
	ErrInvalidStatus    = errors.New("invalid word status")
	ErrMeaningTooLong   = errors.New("meaning exceeds 2000 characters")
	ErrWordNotFound     = errors.New("word not found")
	ErrWordConflict     = errors.New("word already exists with the same banjar, dialect and homonym number")
	ErrWordDeleted      = errors.New("word has been deleted")
)
