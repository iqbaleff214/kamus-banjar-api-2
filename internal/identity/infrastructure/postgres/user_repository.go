package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/postgres/sqlc"
)

type PostgresUserRepository struct {
	queries *q.Queries
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresUserRepository{queries: q.New(db)}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.queries.CreateUser(ctx, q.CreateUserParams{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		Role:            q.UserRole(user.Role),
		IsActive:        user.IsActive,
		EmailVerifiedAt: toNullTime(user.EmailVerifiedAt),
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailConflict
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomain(row), nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomain(row), nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.queries.UpdateUser(ctx, q.UpdateUserParams{
		ID:              user.ID,
		Name:            user.Name,
		PasswordHash:    user.PasswordHash,
		Role:            q.UserRole(user.Role),
		IsActive:        user.IsActive,
		EmailVerifiedAt: toNullTime(user.EmailVerifiedAt),
		UpdatedAt:       time.Now().UTC(),
	})
	return err
}

func toDomain(u q.User) *domain.User {
	du := &domain.User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         domain.Role(u.Role),
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	if u.EmailVerifiedAt.Valid {
		t := u.EmailVerifiedAt.Time
		du.EmailVerifiedAt = &t
	}
	return du
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func isUniqueViolation(err error) bool {
	// pgx wraps pq errors; check for "23505" code
	return err != nil && (containsStr(err.Error(), "23505") || containsStr(err.Error(), "unique"))
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstr(s, sub))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
