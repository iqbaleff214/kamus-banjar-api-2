package domain

import (
	"context"

	"github.com/google/uuid"
)

type ContributionFilter struct {
	Status *ContributionStatus
	Type   *ContributionType
}

type ContributionRepository interface {
	Create(ctx context.Context, c *Contribution) error
	FindByID(ctx context.Context, id uuid.UUID) (*Contribution, error)
	FindByContributor(ctx context.Context, userID uuid.UUID, filter ContributionFilter, page, perPage int) ([]*Contribution, int, error)
	FindAll(ctx context.Context, filter ContributionFilter, page, perPage int) ([]*Contribution, int, error)
	Update(ctx context.Context, c *Contribution) error
}

type VoteRepository interface {
	Upsert(ctx context.Context, v *Vote) (*Vote, error)
	Delete(ctx context.Context, userID uuid.UUID, targetType VoteTargetType, targetID uuid.UUID) error
	FindByUserAndTarget(ctx context.Context, userID uuid.UUID, targetType VoteTargetType, targetID uuid.UUID) (*Vote, error)
}

type BookmarkRepository interface {
	Create(ctx context.Context, b *Bookmark) (*Bookmark, error)
	Delete(ctx context.Context, userID, wordID uuid.UUID) error
	FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*Bookmark, int, error)
	Exists(ctx context.Context, userID, wordID uuid.UUID) (bool, error)
}

type CommentRepository interface {
	Create(ctx context.Context, c *Comment) (*Comment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Comment, error)
	FindByTarget(ctx context.Context, targetType CommentTargetType, targetID uuid.UUID, page, perPage int) ([]*Comment, int, error)
	Update(ctx context.Context, c *Comment) (*Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
