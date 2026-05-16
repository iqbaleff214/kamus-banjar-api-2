package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// WordExistChecker checks if a word exists in the dictionary.
type WordExistChecker struct{ db *sql.DB }

func NewWordExistChecker(pool *pgxpool.Pool) *WordExistChecker {
	return &WordExistChecker{db: stdlib.OpenDBFromPool(pool)}
}

func (c *WordExistChecker) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM words WHERE id = $1 AND deleted_at IS NULL)`, id,
	).Scan(&exists)
	return exists, err
}

// DefinitionExistChecker checks if a definition exists in the dictionary.
type DefinitionExistChecker struct{ db *sql.DB }

func NewDefinitionExistChecker(pool *pgxpool.Pool) *DefinitionExistChecker {
	return &DefinitionExistChecker{db: stdlib.OpenDBFromPool(pool)}
}

func (c *DefinitionExistChecker) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM definitions WHERE id = $1)`, id,
	).Scan(&exists)
	return exists, err
}

// UserEmailVerifier checks whether a user's email is verified.
type UserEmailVerifier struct{ db *sql.DB }

func NewUserEmailVerifier(pool *pgxpool.Pool) *UserEmailVerifier {
	return &UserEmailVerifier{db: stdlib.OpenDBFromPool(pool)}
}

func (v *UserEmailVerifier) IsEmailVerified(ctx context.Context, userID uuid.UUID) (bool, error) {
	var verifiedAt sql.NullTime
	err := v.db.QueryRowContext(ctx,
		`SELECT email_verified_at FROM users WHERE id = $1`, userID,
	).Scan(&verifiedAt)
	if err != nil {
		return false, err
	}
	return verifiedAt.Valid, nil
}
