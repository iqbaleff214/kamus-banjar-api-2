package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/infrastructure/postgres/sqlc"
)

type PostgresCommentRepository struct {
	queries *q.Queries
}

func NewPostgresCommentRepository(pool *pgxpool.Pool) *PostgresCommentRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresCommentRepository{queries: q.New(db)}
}

func (r *PostgresCommentRepository) Create(ctx context.Context, c *domain.Comment) (*domain.Comment, error) {
	row, err := r.queries.CreateComment(ctx, q.CreateCommentParams{
		ID:         c.ID,
		UserID:     c.UserID,
		TargetType: q.CommentTargetType(c.TargetType),
		TargetID:   c.TargetID,
		Body:       c.Body,
	})
	if err != nil {
		return nil, err
	}
	return rowToComment(row), nil
}

func (r *PostgresCommentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	row, err := r.queries.GetCommentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}
	return rowToComment(row), nil
}

func (r *PostgresCommentRepository) FindByTarget(ctx context.Context, targetType domain.CommentTargetType, targetID uuid.UUID, page, perPage int) ([]*domain.Comment, int, error) {
	offset := int32((page - 1) * perPage)

	total, err := r.queries.CountCommentsByTarget(ctx, q.CountCommentsByTargetParams{
		TargetType: q.CommentTargetType(targetType),
		TargetID:   targetID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListCommentsByTarget(ctx, q.ListCommentsByTargetParams{
		TargetType: q.CommentTargetType(targetType),
		TargetID:   targetID,
		Limit:      int32(perPage),
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]*domain.Comment, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowToComment(row))
	}
	return out, int(total), nil
}

func (r *PostgresCommentRepository) Update(ctx context.Context, c *domain.Comment) (*domain.Comment, error) {
	row, err := r.queries.UpdateComment(ctx, q.UpdateCommentParams{
		ID:   c.ID,
		Body: c.Body,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}
	return rowToComment(row), nil
}

func (r *PostgresCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteComment(ctx, id)
}

func rowToComment(row q.Comment) *domain.Comment {
	return &domain.Comment{
		ID:         row.ID,
		UserID:     row.UserID,
		TargetType: domain.CommentTargetType(row.TargetType),
		TargetID:   row.TargetID,
		Body:       row.Body,
		IsFlagged:  row.IsFlagged,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
