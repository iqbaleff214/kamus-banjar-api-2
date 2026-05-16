package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	repo "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/testutil"
)

func newAIRequest(wordID uuid.UUID) *domain.AIRequest {
	return domain.NewAIRequest(
		domain.AIRequestTypeEnrichDefinition,
		&wordID,
		nil,
		uuid.New(),
		"test-model",
		"test prompt",
	)
}

func TestAIRequestRepository_CreateAndFind(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresAIRequestRepository(pool)
	ctx := context.Background()

	req := newAIRequest(uuid.New())
	require.NoError(t, r.Create(ctx, req))

	found, err := r.FindByID(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, req.ID, found.ID)
	assert.Equal(t, domain.AIRequestStatusPending, found.Status)
}

func TestAIRequestRepository_FindByID_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresAIRequestRepository(pool)
	ctx := context.Background()

	_, err := r.FindByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestAIRequestRepository_Update(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresAIRequestRepository(pool)
	ctx := context.Background()

	req := newAIRequest(uuid.New())
	require.NoError(t, r.Create(ctx, req))

	require.NoError(t, req.MarkCompleted(map[string]any{}, map[string]any{"result": "ok"}))
	require.NoError(t, r.Update(ctx, req))

	found, err := r.FindByID(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.AIRequestStatusCompleted, found.Status)
}

func TestAIRequestRepository_ListByWord(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresAIRequestRepository(pool)
	ctx := context.Background()

	wordID := uuid.New()
	req := newAIRequest(wordID)
	require.NoError(t, r.Create(ctx, req))

	items, total, err := r.ListByWord(ctx, wordID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
}

func TestAIRequestRepository_ListPendingReview(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresAIRequestRepository(pool)
	ctx := context.Background()

	req := newAIRequest(uuid.New())
	require.NoError(t, r.Create(ctx, req))
	require.NoError(t, req.MarkCompleted(map[string]any{}, map[string]any{}))
	require.NoError(t, r.Update(ctx, req))

	items, total, err := r.ListPendingReview(ctx, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, items)
}
