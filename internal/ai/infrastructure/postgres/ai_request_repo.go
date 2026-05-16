package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sqlc-dev/pqtype"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/infrastructure/postgres/sqlc"
)

type PostgresAIRequestRepository struct {
	queries *q.Queries
}

func NewPostgresAIRequestRepository(pool *pgxpool.Pool) *PostgresAIRequestRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresAIRequestRepository{queries: q.New(db)}
}

func (r *PostgresAIRequestRepository) Create(ctx context.Context, req *domain.AIRequest) error {
	return r.queries.CreateAIRequest(ctx, q.CreateAIRequestParams{
		ID:                   req.ID,
		Type:                 string(req.Type),
		TargetWordID:         toNullUUID(req.TargetWordID),
		TargetContributionID: toNullUUID(req.TargetContributionID),
		RequestedBy:          req.RequestedBy,
		Model:                req.Model,
		Prompt:               req.Prompt,
		Response:             toNullRaw(req.Response),
		ParsedOutput:         toNullRaw(req.ParsedOutput),
		Status:               string(req.Status),
		ReviewStatus:         string(req.ReviewStatus),
		ReviewedBy:           toNullUUID(req.ReviewedBy),
		ReviewedAt:           toNullTime(req.ReviewedAt),
		CreatedAt:            req.CreatedAt,
	})
}

func (r *PostgresAIRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.AIRequest, error) {
	row, err := r.queries.GetAIRequestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomain(row), nil
}

func (r *PostgresAIRequestRepository) Update(ctx context.Context, req *domain.AIRequest) error {
	return r.queries.UpdateAIRequest(ctx, q.UpdateAIRequestParams{
		ID:           req.ID,
		Response:     toNullRaw(req.Response),
		ParsedOutput: toNullRaw(req.ParsedOutput),
		Status:       string(req.Status),
		ReviewStatus: string(req.ReviewStatus),
		ReviewedBy:   toNullUUID(req.ReviewedBy),
		ReviewedAt:   toNullTime(req.ReviewedAt),
	})
}

func (r *PostgresAIRequestRepository) ListByWord(ctx context.Context, wordID uuid.UUID, page, perPage int) ([]*domain.AIRequest, int, error) {
	nid := uuid.NullUUID{UUID: wordID, Valid: true}
	total, err := r.queries.CountAIRequestsByWord(ctx, nid)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListAIRequestsByWord(ctx, q.ListAIRequestsByWordParams{
		TargetWordID: nid,
		Limit:        int32(perPage),
		Offset:       int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domain.AIRequest, len(rows))
	for i, row := range rows {
		out[i] = toDomain(row)
	}
	return out, int(total), nil
}

func (r *PostgresAIRequestRepository) ListPendingReview(ctx context.Context, page, perPage int) ([]*domain.AIRequest, int, error) {
	total, err := r.queries.CountAIRequestsPendingReview(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListAIRequestsPendingReview(ctx, q.ListAIRequestsPendingReviewParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domain.AIRequest, len(rows))
	for i, row := range rows {
		out[i] = toDomain(row)
	}
	return out, int(total), nil
}

func toDomain(row q.AiRequest) *domain.AIRequest {
	req := &domain.AIRequest{
		ID:           row.ID,
		Type:         domain.AIRequestType(row.Type),
		RequestedBy:  row.RequestedBy,
		Model:        row.Model,
		Prompt:       row.Prompt,
		Status:       domain.AIRequestStatus(row.Status),
		ReviewStatus: domain.AIReviewStatus(row.ReviewStatus),
		CreatedAt:    row.CreatedAt,
	}
	if row.TargetWordID.Valid {
		id := row.TargetWordID.UUID
		req.TargetWordID = &id
	}
	if row.TargetContributionID.Valid {
		id := row.TargetContributionID.UUID
		req.TargetContributionID = &id
	}
	if row.ReviewedBy.Valid {
		id := row.ReviewedBy.UUID
		req.ReviewedBy = &id
	}
	if row.ReviewedAt.Valid {
		req.ReviewedAt = &row.ReviewedAt.Time
	}
	if row.Response.RawMessage != nil {
		_ = json.Unmarshal(row.Response.RawMessage, &req.Response)
	}
	if row.ParsedOutput.RawMessage != nil {
		_ = json.Unmarshal(row.ParsedOutput.RawMessage, &req.ParsedOutput)
	}
	return req
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func toNullRaw(m map[string]any) pqtype.NullRawMessage {
	if m == nil {
		return pqtype.NullRawMessage{}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return pqtype.NullRawMessage{}
	}
	return pqtype.NullRawMessage{RawMessage: b, Valid: true}
}
