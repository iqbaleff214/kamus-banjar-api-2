package commands_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/application/commands"
	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

// ─── fakes ────────────────────────────────────────────────────────────────────

type fakeContribRepo struct {
	store map[uuid.UUID]*domain.Contribution
}

func newFakeContribRepo() *fakeContribRepo {
	return &fakeContribRepo{store: make(map[uuid.UUID]*domain.Contribution)}
}

func (r *fakeContribRepo) Create(_ context.Context, c *domain.Contribution) error {
	r.store[c.ID] = c
	return nil
}
func (r *fakeContribRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Contribution, error) {
	c, ok := r.store[id]
	if !ok {
		return nil, domain.ErrContributionNotFound
	}
	return c, nil
}
func (r *fakeContribRepo) FindByContributor(_ context.Context, userID uuid.UUID, _ domain.ContributionFilter, _, _ int) ([]*domain.Contribution, int, error) {
	var out []*domain.Contribution
	for _, c := range r.store {
		if c.ContributorID == userID {
			out = append(out, c)
		}
	}
	return out, len(out), nil
}
func (r *fakeContribRepo) FindAll(_ context.Context, _ domain.ContributionFilter, _, _ int) ([]*domain.Contribution, int, error) {
	var out []*domain.Contribution
	for _, c := range r.store {
		out = append(out, c)
	}
	return out, len(out), nil
}
func (r *fakeContribRepo) Update(_ context.Context, c *domain.Contribution) error {
	r.store[c.ID] = c
	return nil
}

type fakeWordChecker struct{ exists bool }

func (c *fakeWordChecker) Exists(_ context.Context, _ uuid.UUID) (bool, error) {
	return c.exists, nil
}

type fakeUserVerifier struct{ verified bool }

func (v *fakeUserVerifier) IsEmailVerified(_ context.Context, _ uuid.UUID) (bool, error) {
	return v.verified, nil
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestSubmitContribution_UnverifiedEmail(t *testing.T) {
	svc := commands.NewContributionService(
		newFakeContribRepo(),
		&fakeWordChecker{exists: true},
		&fakeUserVerifier{verified: false},
	)
	_, err := svc.SubmitContribution(context.Background(), uuid.New(), commands.ContributionInput{
		Type:    domain.ContributionTypeNewWord,
		Payload: map[string]any{"banjar": "test"},
	})
	assert.ErrorIs(t, err, commands.ErrEmailNotVerified)
}

func TestSubmitContribution_InvalidTargetWord(t *testing.T) {
	wordID := uuid.New()
	svc := commands.NewContributionService(
		newFakeContribRepo(),
		&fakeWordChecker{exists: false},
		&fakeUserVerifier{verified: true},
	)
	_, err := svc.SubmitContribution(context.Background(), uuid.New(), commands.ContributionInput{
		Type:         domain.ContributionTypeNewDefinition,
		TargetWordID: &wordID,
		Payload:      map[string]any{"meaning": "test"},
	})
	assert.ErrorIs(t, err, commands.ErrTargetWordNotFound)
}

func TestWithdrawContribution_Success(t *testing.T) {
	repo := newFakeContribRepo()
	svc := commands.NewContributionService(repo, &fakeWordChecker{exists: true}, &fakeUserVerifier{verified: true})

	userID := uuid.New()
	wordID := uuid.New()
	c, err := svc.SubmitContribution(context.Background(), userID, commands.ContributionInput{
		Type:         domain.ContributionTypeNewDefinition,
		TargetWordID: &wordID,
		Payload:      map[string]any{},
	})
	require.NoError(t, err)

	result, err := svc.WithdrawContribution(context.Background(), userID, c.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ContributionStatusWithdrawn, result.Status)
}
