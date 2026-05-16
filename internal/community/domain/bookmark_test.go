package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

func TestBookmark_Create(t *testing.T) {
	userID := uuid.New()
	wordID := uuid.New()
	b := domain.NewBookmark(userID, wordID)
	assert.Equal(t, userID, b.UserID)
	assert.Equal(t, wordID, b.WordID)
	assert.NotEqual(t, uuid.Nil, b.ID)
}
