package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository is the persistence interface for the User aggregate.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}
