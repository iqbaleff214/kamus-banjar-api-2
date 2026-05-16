package commands

import (
	"context"

	"github.com/google/uuid"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

// ModerationRepository provides moderation-specific data queries and audit logging.
type ModerationRepository interface {
	GetPendingContributions(ctx context.Context, ctype *communitydomain.ContributionType, page, perPage int) ([]*communitydomain.Contribution, int, error)
	GetFlaggedComments(ctx context.Context, page, perPage int) ([]*communitydomain.Comment, int, error)
	GetStats(ctx context.Context) (*moderationdomain.ModerationStats, error)
	CreateAuditLog(ctx context.Context, log *moderationdomain.AuditLog) error
	ListUsers(ctx context.Context, role, query string, isActive *bool, page, perPage int) ([]*identitydomain.User, int, error)
}

// ContributionWriter loads and persists contributions for moderation actions.
type ContributionWriter interface {
	FindByID(ctx context.Context, id uuid.UUID) (*communitydomain.Contribution, error)
	Update(ctx context.Context, c *communitydomain.Contribution) error
}

// WordWriter provides read/write access to the dictionary for merging approved contributions.
type WordWriter interface {
	FindByID(ctx context.Context, id uuid.UUID) (*dictdomain.Word, error)
	Create(ctx context.Context, word *dictdomain.Word) error
	Update(ctx context.Context, word *dictdomain.Word) error
}

// UserWriter provides read/write access to users for ban/unban/role-change operations.
type UserWriter interface {
	FindByID(ctx context.Context, id uuid.UUID) (*identitydomain.User, error)
	Update(ctx context.Context, user *identitydomain.User) error
}
