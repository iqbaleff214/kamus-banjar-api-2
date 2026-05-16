package domain_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

func TestNewComment_BodyTooLong(t *testing.T) {
	body := strings.Repeat("a", 1001)
	_, err := domain.NewComment(uuid.New(), domain.CommentTargetWord, uuid.New(), body)
	assert.ErrorIs(t, err, domain.ErrBodyTooLong)
}

func TestNewComment_EmptyBody(t *testing.T) {
	_, err := domain.NewComment(uuid.New(), domain.CommentTargetWord, uuid.New(), "")
	assert.ErrorIs(t, err, domain.ErrEmptyBody)
}

func TestComment_EditByNonOwner(t *testing.T) {
	c := newTestComment(t)
	err := c.Edit(uuid.New(), "new body")
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestComment_Flag(t *testing.T) {
	c := newTestComment(t)
	assert.False(t, c.IsFlagged)
	c.Flag()
	assert.True(t, c.IsFlagged)
	// idempotent
	c.Flag()
	assert.True(t, c.IsFlagged)
}

func newTestComment(t *testing.T) *domain.Comment {
	t.Helper()
	c, err := domain.NewComment(uuid.New(), domain.CommentTargetWord, uuid.New(), "good word")
	require.NoError(t, err)
	return c
}
