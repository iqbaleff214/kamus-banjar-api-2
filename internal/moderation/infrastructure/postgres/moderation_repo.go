package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sqlc-dev/pqtype"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	communityq "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/infrastructure/postgres/sqlc"
	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/infrastructure/postgres/sqlc"
)

type PostgresModerationRepository struct {
	modQ       *q.Queries
	communityQ *communityq.Queries
}

func NewPostgresModerationRepository(pool *pgxpool.Pool) *PostgresModerationRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresModerationRepository{
		modQ:       q.New(db),
		communityQ: communityq.New(db),
	}
}

func (r *PostgresModerationRepository) GetPendingContributions(ctx context.Context, ctype *communitydomain.ContributionType, page, perPage int) ([]*communitydomain.Contribution, int, error) {
	statusFilter := string(communitydomain.ContributionStatusPending)
	typeFilter := ""
	if ctype != nil {
		typeFilter = string(*ctype)
	}
	offset := int32((page - 1) * perPage)

	total, err := r.communityQ.CountAllContributions(ctx, communityq.CountAllContributionsParams{
		Column1: statusFilter,
		Column2: typeFilter,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.communityQ.ListAllContributions(ctx, communityq.ListAllContributionsParams{
		Column1: statusFilter,
		Column2: typeFilter,
		Limit:   int32(perPage),
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]*communitydomain.Contribution, 0, len(rows))
	for _, row := range rows {
		c, err := rowToContribution(row)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, int(total), nil
}

func (r *PostgresModerationRepository) GetFlaggedComments(ctx context.Context, page, perPage int) ([]*communitydomain.Comment, int, error) {
	offset := int32((page - 1) * perPage)

	total, err := r.communityQ.CountFlaggedComments(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.communityQ.ListFlaggedComments(ctx, communityq.ListFlaggedCommentsParams{
		Limit:  int32(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]*communitydomain.Comment, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowToComment(row))
	}
	return out, int(total), nil
}

func (r *PostgresModerationRepository) GetStats(ctx context.Context) (*moderationdomain.ModerationStats, error) {
	row, err := r.modQ.GetModerationStats(ctx)
	if err != nil {
		return nil, err
	}
	return &moderationdomain.ModerationStats{
		PendingContributions: int(row.PendingContributions),
		FlaggedComments:      int(row.FlaggedComments),
		ApprovedThisWeek:     int(row.ApprovedThisWeek),
		RejectedThisWeek:     int(row.RejectedThisWeek),
	}, nil
}

func (r *PostgresModerationRepository) CreateAuditLog(ctx context.Context, log *moderationdomain.AuditLog) error {
	var metadata pqtype.NullRawMessage
	if log.Metadata != nil {
		b, err := json.Marshal(log.Metadata)
		if err != nil {
			return err
		}
		metadata = pqtype.NullRawMessage{RawMessage: b, Valid: true}
	}
	return r.modQ.CreateAuditLog(ctx, q.CreateAuditLogParams{
		ID:         log.ID,
		ActorID:    log.ActorID,
		Action:     string(log.Action),
		TargetType: log.TargetType,
		TargetID:   log.TargetID,
		Metadata:   metadata,
		CreatedAt:  log.CreatedAt,
	})
}

func (r *PostgresModerationRepository) ListUsers(ctx context.Context, role, query string, isActive *bool, page, perPage int) ([]*identitydomain.User, int, error) {
	activeFilter := ""
	if isActive != nil {
		if *isActive {
			activeFilter = "true"
		} else {
			activeFilter = "false"
		}
	}
	offset := int32((page - 1) * perPage)

	total, err := r.modQ.CountUsersAdmin(ctx, q.CountUsersAdminParams{
		Column1: role,
		Column2: activeFilter,
		Column3: query,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.modQ.ListUsersAdmin(ctx, q.ListUsersAdminParams{
		Column1: role,
		Column2: activeFilter,
		Column3: query,
		Limit:   int32(perPage),
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]*identitydomain.User, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowToUser(row))
	}
	return out, int(total), nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func rowToContribution(row communityq.Contribution) (*communitydomain.Contribution, error) {
	var payload map[string]any
	if err := json.Unmarshal(row.Payload, &payload); err != nil {
		payload = make(map[string]any)
	}
	c := &communitydomain.Contribution{
		ID:            row.ID,
		Type:          communitydomain.ContributionType(row.Type),
		ContributorID: row.ContributorID,
		Payload:       payload,
		Status:        communitydomain.ContributionStatus(row.Status),
		SubmittedAt:   row.SubmittedAt,
	}
	if row.TargetWordID.Valid {
		uid := row.TargetWordID.UUID
		c.TargetWordID = &uid
	}
	if row.ReviewerID.Valid {
		uid := row.ReviewerID.UUID
		c.ReviewerID = &uid
	}
	if row.ReviewerNote.Valid {
		c.ReviewerNote = &row.ReviewerNote.String
	}
	if row.ReviewedAt.Valid {
		t := row.ReviewedAt.Time
		c.ReviewedAt = &t
	}
	return c, nil
}

func rowToComment(row communityq.Comment) *communitydomain.Comment {
	return &communitydomain.Comment{
		ID:         row.ID,
		UserID:     row.UserID,
		TargetType: communitydomain.CommentTargetType(row.TargetType),
		TargetID:   row.TargetID,
		Body:       row.Body,
		IsFlagged:  row.IsFlagged,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func rowToUser(row q.User) *identitydomain.User {
	u := &identitydomain.User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         identitydomain.Role(row.Role),
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.EmailVerifiedAt.Valid {
		t := row.EmailVerifiedAt.Time
		u.EmailVerifiedAt = &t
	}
	return u
}

// silence unused import — time is used for AuditLog.CreatedAt
var _ = time.Now
var _ = uuid.Nil
