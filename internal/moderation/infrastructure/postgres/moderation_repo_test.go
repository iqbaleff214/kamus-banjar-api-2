package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
	repo "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/testutil"
)

func TestModerationRepository_GetStats(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresModerationRepository(pool)
	ctx := context.Background()

	stats, err := r.GetStats(ctx)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestModerationRepository_GetPendingContributions(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresModerationRepository(pool)
	ctx := context.Background()

	items, total, err := r.GetPendingContributions(ctx, nil, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 0)
	assert.NotNil(t, items)
}

func TestModerationRepository_GetFlaggedComments(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresModerationRepository(pool)
	ctx := context.Background()

	items, total, err := r.GetFlaggedComments(ctx, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 0)
	assert.NotNil(t, items)
}

func TestModerationRepository_CreateAuditLog(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresModerationRepository(pool)
	ctx := context.Background()

	log, err := moderationdomain.NewAuditLog(uuid.New(), moderationdomain.ActionBanUser, "user", uuid.New(), nil)
	require.NoError(t, err)

	require.NoError(t, r.CreateAuditLog(ctx, log))
}

func TestModerationRepository_ListUsers(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresModerationRepository(pool)
	ctx := context.Background()

	users, total, err := r.ListUsers(ctx, "", "", nil, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 0)
	assert.NotNil(t, users)
}
