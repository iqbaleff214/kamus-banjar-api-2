package domain

import "github.com/google/uuid"

const MaxMeaningLength = 2000

type Definition struct {
	ID        uuid.UUID
	Meaning   string
	SortOrder int
	Source    Source
	Upvotes   int
	Downvotes int
}

func NewDefinition(meaning string, sortOrder int, source Source) (*Definition, error) {
	if len(meaning) > MaxMeaningLength {
		return nil, ErrMeaningTooLong
	}
	return &Definition{
		ID:        uuid.New(),
		Meaning:   meaning,
		SortOrder: sortOrder,
		Source:    source,
	}, nil
}

func (d *Definition) NetScore() int {
	return d.Upvotes - d.Downvotes
}
