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

type PostgresVoteRepository struct {
	queries *q.Queries
}

func NewPostgresVoteRepository(pool *pgxpool.Pool) *PostgresVoteRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresVoteRepository{queries: q.New(db)}
}

func (r *PostgresVoteRepository) Upsert(ctx context.Context, v *domain.Vote) (*domain.Vote, error) {
	row, err := r.queries.UpsertVote(ctx, q.UpsertVoteParams{
		ID:         v.ID,
		UserID:     v.UserID,
		TargetType: q.VoteTargetType(v.TargetType),
		TargetID:   v.TargetID,
		Value:      q.VoteValue(v.Value),
	})
	if err != nil {
		return nil, err
	}
	return &domain.Vote{
		ID:         row.ID,
		UserID:     row.UserID,
		TargetType: domain.VoteTargetType(row.TargetType),
		TargetID:   row.TargetID,
		Value:      domain.VoteValue(row.Value),
		CreatedAt:  row.CreatedAt,
	}, nil
}

func (r *PostgresVoteRepository) Delete(ctx context.Context, userID uuid.UUID, targetType domain.VoteTargetType, targetID uuid.UUID) error {
	return r.queries.DeleteVote(ctx, q.DeleteVoteParams{
		UserID:     userID,
		TargetType: q.VoteTargetType(targetType),
		TargetID:   targetID,
	})
}

func (r *PostgresVoteRepository) FindByUserAndTarget(ctx context.Context, userID uuid.UUID, targetType domain.VoteTargetType, targetID uuid.UUID) (*domain.Vote, error) {
	row, err := r.queries.GetVoteByUserAndTarget(ctx, q.GetVoteByUserAndTargetParams{
		UserID:     userID,
		TargetType: q.VoteTargetType(targetType),
		TargetID:   targetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrVoteNotFound
		}
		return nil, err
	}
	return &domain.Vote{
		ID:         row.ID,
		UserID:     row.UserID,
		TargetType: domain.VoteTargetType(row.TargetType),
		TargetID:   row.TargetID,
		Value:      domain.VoteValue(row.Value),
		CreatedAt:  row.CreatedAt,
	}, nil
}
