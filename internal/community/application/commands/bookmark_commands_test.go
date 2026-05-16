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

type fakeBookmarkRepo struct {
	store map[string]*domain.Bookmark // key: userID+wordID
}

func newFakeBookmarkRepo() *fakeBookmarkRepo {
	return &fakeBookmarkRepo{store: make(map[string]*domain.Bookmark)}
}

func bmKey(userID, wordID uuid.UUID) string { return userID.String() + wordID.String() }

func (r *fakeBookmarkRepo) Create(_ context.Context, b *domain.Bookmark) (*domain.Bookmark, error) {
	k := bmKey(b.UserID, b.WordID)
	if _, exists := r.store[k]; exists {
		return nil, domain.ErrBookmarkConflict
	}
	r.store[k] = b
	return b, nil
}
func (r *fakeBookmarkRepo) Delete(_ context.Context, userID, wordID uuid.UUID) error {
	delete(r.store, bmKey(userID, wordID))
	return nil
}
func (r *fakeBookmarkRepo) FindByUser(_ context.Context, userID uuid.UUID, _, _ int) ([]*domain.Bookmark, int, error) {
	var out []*domain.Bookmark
	for _, b := range r.store {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	return out, len(out), nil
}
func (r *fakeBookmarkRepo) Exists(_ context.Context, userID, wordID uuid.UUID) (bool, error) {
	_, ok := r.store[bmKey(userID, wordID)]
	return ok, nil
}

func TestAddBookmark_Duplicate(t *testing.T) {
	repo := newFakeBookmarkRepo()
	svc := commands.NewBookmarkService(repo, &fakeWordChecker{exists: true})
	userID, wordID := uuid.New(), uuid.New()

	_, err := svc.AddBookmark(context.Background(), userID, wordID)
	require.NoError(t, err)

	_, err = svc.AddBookmark(context.Background(), userID, wordID)
	assert.ErrorIs(t, err, domain.ErrBookmarkConflict)
}

func TestRemoveBookmark_NotFound(t *testing.T) {
	svc := commands.NewBookmarkService(newFakeBookmarkRepo(), &fakeWordChecker{exists: true})
	err := svc.RemoveBookmark(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrBookmarkNotFound)
}
