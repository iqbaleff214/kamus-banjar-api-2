package commands

import (
	"context"

	"github.com/google/uuid"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

type WordAccessor interface {
	FindByID(ctx context.Context, id uuid.UUID) (*dictdomain.Word, error)
	FindByBanjar(ctx context.Context, banjar string) (*dictdomain.Word, error)
	Update(ctx context.Context, word *dictdomain.Word) error
}

type ContributionReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*communitydomain.Contribution, error)
}

type AuditLogWriter interface {
	CreateAuditLog(ctx context.Context, log *moderationdomain.AuditLog) error
}
