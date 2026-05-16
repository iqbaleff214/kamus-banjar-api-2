package commands_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/application/commands"
	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCommentRepo struct {
	store map[uuid.UUID]*domain.Comment
}

func newFakeCommentRepo() *fakeCommentRepo {
	return &fakeCommentRepo{store: make(map[uuid.UUID]*domain.Comment)}
}

func (r *fakeCommentRepo) Create(_ context.Context, c *domain.Comment) (*domain.Comment, error) {
	r.store[c.ID] = c
	return c, nil
}
func (r *fakeCommentRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	c, ok := r.store[id]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	return c, nil
}
func (r *fakeCommentRepo) FindByTarget(_ context.Context, _ domain.CommentTargetType, _ uuid.UUID, _, _ int) ([]*domain.Comment, int, error) {
	return nil, 0, nil
}
func (r *fakeCommentRepo) Update(_ context.Context, c *domain.Comment) (*domain.Comment, error) {
	r.store[c.ID] = c
	return c, nil
}
func (r *fakeCommentRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.store, id)
	return nil
}

func TestEditComment_NonOwner(t *testing.T) {
	repo := newFakeCommentRepo()
	svc := commands.NewCommentService(repo, &fakeWordChecker{exists: true})

	ownerID := uuid.New()
	c, err := svc.PostComment(context.Background(), ownerID, uuid.New(), "hello")
	require.NoError(t, err)

	_, err = svc.EditComment(context.Background(), uuid.New(), c.ID, "different owner")
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDeleteComment_AdminCanDeleteAny(t *testing.T) {
	repo := newFakeCommentRepo()
	svc := commands.NewCommentService(repo, &fakeWordChecker{exists: true})

	ownerID := uuid.New()
	c, err := svc.PostComment(context.Background(), ownerID, uuid.New(), "hello")
	require.NoError(t, err)

	err = svc.DeleteComment(context.Background(), uuid.New(), "admin", c.ID)
	require.NoError(t, err)
}

func TestFlagComment_Idempotent(t *testing.T) {
	repo := newFakeCommentRepo()
	svc := commands.NewCommentService(repo, &fakeWordChecker{exists: true})

	c, err := svc.PostComment(context.Background(), uuid.New(), uuid.New(), "hello")
	require.NoError(t, err)

	c1, err := svc.FlagComment(context.Background(), c.ID)
	require.NoError(t, err)
	assert.True(t, c1.IsFlagged)

	c2, err := svc.FlagComment(context.Background(), c.ID)
	require.NoError(t, err)
	assert.True(t, c2.IsFlagged)
}
