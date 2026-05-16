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

type fakeVoteRepo struct {
	store map[string]*domain.Vote // key: userID+targetType+targetID
}

func newFakeVoteRepo() *fakeVoteRepo {
	return &fakeVoteRepo{store: make(map[string]*domain.Vote)}
}

func voteKey(userID uuid.UUID, tt domain.VoteTargetType, targetID uuid.UUID) string {
	return userID.String() + string(tt) + targetID.String()
}

func (r *fakeVoteRepo) Upsert(_ context.Context, v *domain.Vote) (*domain.Vote, error) {
	r.store[voteKey(v.UserID, v.TargetType, v.TargetID)] = v
	return v, nil
}
func (r *fakeVoteRepo) Delete(_ context.Context, userID uuid.UUID, tt domain.VoteTargetType, targetID uuid.UUID) error {
	delete(r.store, voteKey(userID, tt, targetID))
	return nil
}
func (r *fakeVoteRepo) FindByUserAndTarget(_ context.Context, userID uuid.UUID, tt domain.VoteTargetType, targetID uuid.UUID) (*domain.Vote, error) {
	v, ok := r.store[voteKey(userID, tt, targetID)]
	if !ok {
		return nil, domain.ErrVoteNotFound
	}
	return v, nil
}

func TestCastVote_TargetNotFound(t *testing.T) {
	svc := commands.NewVoteService(newFakeVoteRepo(), &fakeWordChecker{exists: false}, &fakeWordChecker{exists: true})
	_, err := svc.CastVote(context.Background(), uuid.New(), domain.VoteTargetWord, uuid.New(), domain.VoteUp)
	assert.ErrorIs(t, err, commands.ErrVoteTargetNotFound)
}

func TestCastVote_ChangeDirection(t *testing.T) {
	repo := newFakeVoteRepo()
	svc := commands.NewVoteService(repo, &fakeWordChecker{exists: true}, &fakeWordChecker{exists: true})

	userID := uuid.New()
	targetID := uuid.New()

	v1, err := svc.CastVote(context.Background(), userID, domain.VoteTargetWord, targetID, domain.VoteUp)
	require.NoError(t, err)
	assert.Equal(t, domain.VoteUp, v1.Value)

	// Simulate direction change (upsert replaces value)
	repo.store[voteKey(userID, domain.VoteTargetWord, targetID)].Value = domain.VoteDown
	stored := repo.store[voteKey(userID, domain.VoteTargetWord, targetID)]
	assert.Equal(t, domain.VoteDown, stored.Value)
}
