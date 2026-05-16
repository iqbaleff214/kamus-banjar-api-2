package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	repo "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/testutil"
)

func newTestUser(t *testing.T) *domain.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("Password1!"), bcrypt.MinCost)
	require.NoError(t, err)
	return &domain.User{
		ID:           uuid.New(),
		Name:         "Test User",
		Email:        uuid.New().String() + "@example.com",
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
		UpdatedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func TestPostgresUserRepository_CreateAndFind(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser(t)
	require.NoError(t, r.Create(ctx, user))

	found, err := r.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestPostgresUserRepository_FindByEmail(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser(t)
	require.NoError(t, r.Create(ctx, user))

	found, err := r.FindByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestPostgresUserRepository_FindByEmail_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresUserRepository(pool)
	ctx := context.Background()

	_, err := r.FindByEmail(ctx, "nobody@example.com")
	assert.Error(t, err)
}

func TestPostgresUserRepository_Update(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser(t)
	require.NoError(t, r.Create(ctx, user))

	user.Name = "Updated Name"
	require.NoError(t, r.Update(ctx, user))

	found, err := r.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
}

func TestPostgresUserRepository_Create_DuplicateEmail(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresUserRepository(pool)
	ctx := context.Background()

	user := newTestUser(t)
	require.NoError(t, r.Create(ctx, user))

	user2 := newTestUser(t)
	user2.Email = user.Email
	err := r.Create(ctx, user2)
	assert.Error(t, err)
}
