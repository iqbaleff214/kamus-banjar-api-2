package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

func TestCastVote_Valid(t *testing.T) {
	v, err := domain.CastVote(uuid.New(), domain.VoteTargetWord, uuid.New(), domain.VoteUp)
	require.NoError(t, err)
	assert.Equal(t, domain.VoteTargetWord, v.TargetType)
	assert.Equal(t, domain.VoteUp, v.Value)
}

func TestCastVote_InvalidTargetType(t *testing.T) {
	_, err := domain.CastVote(uuid.New(), "comment", uuid.New(), domain.VoteUp)
	assert.ErrorIs(t, err, domain.ErrInvalidVoteTarget)
}

func TestCastVote_InvalidValue(t *testing.T) {
	_, err := domain.CastVote(uuid.New(), domain.VoteTargetWord, uuid.New(), "neutral")
	assert.ErrorIs(t, err, domain.ErrInvalidVoteValue)
}
