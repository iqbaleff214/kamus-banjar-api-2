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

type PostgresBookmarkRepository struct {
	queries *q.Queries
}

func NewPostgresBookmarkRepository(pool *pgxpool.Pool) *PostgresBookmarkRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresBookmarkRepository{queries: q.New(db)}
}

func (r *PostgresBookmarkRepository) Create(ctx context.Context, b *domain.Bookmark) (*domain.Bookmark, error) {
	row, err := r.queries.CreateBookmark(ctx, q.CreateBookmarkParams{
		ID:     b.ID,
		UserID: b.UserID,
		WordID: b.WordID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrBookmarkConflict
		}
		return nil, err
	}
	return &domain.Bookmark{
		ID:        row.ID,
		UserID:    row.UserID,
		WordID:    row.WordID,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *PostgresBookmarkRepository) Delete(ctx context.Context, userID, wordID uuid.UUID) error {
	return r.queries.DeleteBookmark(ctx, q.DeleteBookmarkParams{
		UserID: userID,
		WordID: wordID,
	})
}

func (r *PostgresBookmarkRepository) FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*domain.Bookmark, int, error) {
	offset := int32((page - 1) * perPage)

	total, err := r.queries.CountBookmarksByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListBookmarksByUser(ctx, q.ListBookmarksByUserParams{
		UserID: userID,
		Limit:  int32(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]*domain.Bookmark, 0, len(rows))
	for _, row := range rows {
		out = append(out, &domain.Bookmark{
			ID:        row.ID,
			UserID:    row.UserID,
			WordID:    row.WordID,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, int(total), nil
}

func (r *PostgresBookmarkRepository) Exists(ctx context.Context, userID, wordID uuid.UUID) (bool, error) {
	_, err := r.queries.GetBookmarkByUserAndWord(ctx, q.GetBookmarkByUserAndWordParams{
		UserID: userID,
		WordID: wordID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
