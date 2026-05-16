package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/infrastructure/postgres/sqlc"
)

type PostgresContributionRepository struct {
	queries *q.Queries
}

func NewPostgresContributionRepository(pool *pgxpool.Pool) *PostgresContributionRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresContributionRepository{queries: q.New(db)}
}

func (r *PostgresContributionRepository) Create(ctx context.Context, c *domain.Contribution) error {
	payload, err := json.Marshal(c.Payload)
	if err != nil {
		return err
	}
	var targetWordID uuid.NullUUID
	if c.TargetWordID != nil {
		targetWordID = uuid.NullUUID{UUID: *c.TargetWordID, Valid: true}
	}
	row, err := r.queries.CreateContribution(ctx, q.CreateContributionParams{
		ID:            c.ID,
		Type:          q.ContributionType(c.Type),
		ContributorID: c.ContributorID,
		TargetWordID:  targetWordID,
		Payload:       payload,
	})
	if err != nil {
		return err
	}
	c.SubmittedAt = row.SubmittedAt
	return nil
}

func (r *PostgresContributionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contribution, error) {
	row, err := r.queries.GetContributionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrContributionNotFound
		}
		return nil, err
	}
	return rowToContribution(row)
}

func (r *PostgresContributionRepository) FindByContributor(ctx context.Context, userID uuid.UUID, filter domain.ContributionFilter, page, perPage int) ([]*domain.Contribution, int, error) {
	statusFilter, typeFilter := filterStrings(filter)
	offset := int32((page - 1) * perPage)

	total, err := r.queries.CountContributionsByContributor(ctx, q.CountContributionsByContributorParams{
		ContributorID: userID,
		Column2:       statusFilter,
		Column3:       typeFilter,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListContributionsByContributor(ctx, q.ListContributionsByContributorParams{
		ContributorID: userID,
		Column2:       statusFilter,
		Column3:       typeFilter,
		Limit:         int32(perPage),
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}
	contribs, _, err := rowsToContributions(rows)
	return contribs, int(total), err
}

func (r *PostgresContributionRepository) FindAll(ctx context.Context, filter domain.ContributionFilter, page, perPage int) ([]*domain.Contribution, int, error) {
	statusFilter, typeFilter := filterStrings(filter)
	offset := int32((page - 1) * perPage)

	total, err := r.queries.CountAllContributions(ctx, q.CountAllContributionsParams{
		Column1: statusFilter,
		Column2: typeFilter,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListAllContributions(ctx, q.ListAllContributionsParams{
		Column1: statusFilter,
		Column2: typeFilter,
		Limit:   int32(perPage),
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}
	contribs, _, err := rowsToContributions(rows)
	return contribs, int(total), err
}

func (r *PostgresContributionRepository) Update(ctx context.Context, c *domain.Contribution) error {
	var reviewerID uuid.NullUUID
	if c.ReviewerID != nil {
		reviewerID = uuid.NullUUID{UUID: *c.ReviewerID, Valid: true}
	}
	var reviewerNote sql.NullString
	if c.ReviewerNote != nil {
		reviewerNote = sql.NullString{String: *c.ReviewerNote, Valid: true}
	}
	var reviewedAt sql.NullTime
	if c.ReviewedAt != nil {
		reviewedAt = sql.NullTime{Time: *c.ReviewedAt, Valid: true}
	}
	_, err := r.queries.UpdateContribution(ctx, q.UpdateContributionParams{
		ID:           c.ID,
		Status:       q.ContributionStatus(c.Status),
		ReviewerID:   reviewerID,
		ReviewerNote: reviewerNote,
		ReviewedAt:   reviewedAt,
	})
	return err
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func rowToContribution(row q.Contribution) (*domain.Contribution, error) {
	var payload map[string]any
	if err := json.Unmarshal(row.Payload, &payload); err != nil {
		payload = make(map[string]any)
	}

	c := &domain.Contribution{
		ID:            row.ID,
		Type:          domain.ContributionType(row.Type),
		ContributorID: row.ContributorID,
		Payload:       payload,
		Status:        domain.ContributionStatus(row.Status),
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

func rowsToContributions(rows []q.Contribution) ([]*domain.Contribution, int, error) {
	out := make([]*domain.Contribution, 0, len(rows))
	for _, row := range rows {
		c, err := rowToContribution(row)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, len(out), nil
}

func filterStrings(f domain.ContributionFilter) (string, string) {
	statusFilter := ""
	if f.Status != nil {
		statusFilter = string(*f.Status)
	}
	typeFilter := ""
	if f.Type != nil {
		typeFilter = string(*f.Type)
	}
	return statusFilter, typeFilter
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}

// silence unused import warning — time is used for ReviewedAt
var _ = time.Now
