package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	repo "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/testutil"
)

func newContrib(userID, wordID uuid.UUID) *domain.Contribution {
	c, _ := domain.NewContribution(userID, domain.ContributionTypeNewDefinition, &wordID, map[string]any{"meaning": "test"})
	return c
}

// ─── ContributionRepository ───────────────────────────────────────────────────

func TestContributionRepository_CreateAndFind(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresContributionRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	wordID := uuid.New()
	c := newContrib(userID, wordID)
	require.NoError(t, r.Create(ctx, c))

	found, err := r.FindByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.Type, found.Type)
}

func TestContributionRepository_FindByID_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresContributionRepository(pool)
	ctx := context.Background()

	_, err := r.FindByID(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrContributionNotFound)
}

func TestContributionRepository_Update(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresContributionRepository(pool)
	ctx := context.Background()

	c := newContrib(uuid.New(), uuid.New())
	require.NoError(t, r.Create(ctx, c))

	reviewerID := uuid.New()
	require.NoError(t, c.Approve(reviewerID, "looks good"))
	require.NoError(t, r.Update(ctx, c))

	found, err := r.FindByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ContributionStatusApproved, found.Status)
}

func TestContributionRepository_FindAll(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresContributionRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	c := newContrib(userID, uuid.New())
	require.NoError(t, r.Create(ctx, c))

	all, total, err := r.FindAll(ctx, domain.ContributionFilter{}, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, all)
}

// ─── BookmarkRepository ───────────────────────────────────────────────────────

func TestBookmarkRepository_CreateAndDelete(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresBookmarkRepository(pool)
	ctx := context.Background()

	userID, wordID := uuid.New(), uuid.New()
	b := &domain.Bookmark{UserID: userID, WordID: wordID, CreatedAt: time.Now()}
	created, err := r.Create(ctx, b)
	require.NoError(t, err)
	assert.Equal(t, userID, created.UserID)

	require.NoError(t, r.Delete(ctx, userID, wordID))

	exists, err := r.Exists(ctx, userID, wordID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestBookmarkRepository_Create_Duplicate(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresBookmarkRepository(pool)
	ctx := context.Background()

	userID, wordID := uuid.New(), uuid.New()
	b := &domain.Bookmark{UserID: userID, WordID: wordID, CreatedAt: time.Now()}
	_, err := r.Create(ctx, b)
	require.NoError(t, err)

	_, err = r.Create(ctx, &domain.Bookmark{UserID: userID, WordID: wordID, CreatedAt: time.Now()})
	assert.ErrorIs(t, err, domain.ErrBookmarkConflict)
}

func TestBookmarkRepository_FindByUser(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresBookmarkRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	_, err := r.Create(ctx, &domain.Bookmark{UserID: userID, WordID: uuid.New(), CreatedAt: time.Now()})
	require.NoError(t, err)

	items, total, err := r.FindByUser(ctx, userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
}

// ─── VoteRepository ───────────────────────────────────────────────────────────

func TestVoteRepository_UpsertAndDelete(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresVoteRepository(pool)
	ctx := context.Background()

	userID, targetID := uuid.New(), uuid.New()
	v := &domain.Vote{
		UserID:     userID,
		TargetType: domain.VoteTargetWord,
		TargetID:   targetID,
		Value:      domain.VoteUp,
	}
	created, err := r.Upsert(ctx, v)
	require.NoError(t, err)
	assert.Equal(t, domain.VoteUp, created.Value)

	found, err := r.FindByUserAndTarget(ctx, userID, domain.VoteTargetWord, targetID)
	require.NoError(t, err)
	assert.Equal(t, domain.VoteUp, found.Value)

	require.NoError(t, r.Delete(ctx, userID, domain.VoteTargetWord, targetID))

	_, err = r.FindByUserAndTarget(ctx, userID, domain.VoteTargetWord, targetID)
	assert.ErrorIs(t, err, domain.ErrVoteNotFound)
}

// ─── CommentRepository ────────────────────────────────────────────────────────

func TestCommentRepository_CreateAndFind(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresCommentRepository(pool)
	ctx := context.Background()

	userID, wordID := uuid.New(), uuid.New()
	c, err := domain.NewComment(userID, domain.CommentTargetWord, wordID, "katanya bagus")
	require.NoError(t, err)

	created, err := r.Create(ctx, c)
	require.NoError(t, err)
	assert.Equal(t, "katanya bagus", created.Body)

	found, err := r.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestCommentRepository_FindByID_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresCommentRepository(pool)
	ctx := context.Background()

	_, err := r.FindByID(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrCommentNotFound)
}

func TestCommentRepository_FindByTarget(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresCommentRepository(pool)
	ctx := context.Background()

	userID, wordID := uuid.New(), uuid.New()
	c, _ := domain.NewComment(userID, domain.CommentTargetWord, wordID, "nice")
	_, err := r.Create(ctx, c)
	require.NoError(t, err)

	items, total, err := r.FindByTarget(ctx, domain.CommentTargetWord, wordID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
}

func TestCommentRepository_UpdateAndDelete(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresCommentRepository(pool)
	ctx := context.Background()

	userID, wordID := uuid.New(), uuid.New()
	c, _ := domain.NewComment(userID, domain.CommentTargetWord, wordID, "original")
	created, err := r.Create(ctx, c)
	require.NoError(t, err)

	require.NoError(t, created.Edit(userID, "updated"))
	_, err = r.Update(ctx, created)
	require.NoError(t, err)

	require.NoError(t, r.Delete(ctx, created.ID))
	_, err = r.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, domain.ErrCommentNotFound)
}
