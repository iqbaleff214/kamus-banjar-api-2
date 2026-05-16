package domain

import (
	"context"

	"github.com/google/uuid"
)

type AIRequestRepository interface {
	Create(ctx context.Context, r *AIRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*AIRequest, error)
	Update(ctx context.Context, r *AIRequest) error
	ListByWord(ctx context.Context, wordID uuid.UUID, page, perPage int) ([]*AIRequest, int, error)
	ListPendingReview(ctx context.Context, page, perPage int) ([]*AIRequest, int, error)
}
