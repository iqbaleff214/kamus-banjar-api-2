package domain

import (
	"context"

	"github.com/google/uuid"
)

// WordFilter defines optional filters for word list/search queries.
type WordFilter struct {
	WordClass *WordClass
	IsRoot    *bool
	Source    *Source
	Status    *WordStatus
	Sort      string // "alphabetical" | "most_voted" | "recently_added"
}

// WordRepository is the persistence interface for the Word aggregate.
type WordRepository interface {
	Create(ctx context.Context, word *Word) error
	FindByID(ctx context.Context, id uuid.UUID) (*Word, error)
	FindAll(ctx context.Context, filter WordFilter, page, perPage int) ([]*Word, int, error)
	Search(ctx context.Context, query string, filter WordFilter, page, perPage int) ([]*Word, int, error)
	Update(ctx context.Context, word *Word) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
