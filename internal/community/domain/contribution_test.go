package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContribution_ApprovePending(t *testing.T) {
	contrib := pendingContribution(t)
	reviewer := uuid.New()
	err := contrib.Approve(reviewer, "looks good")
	require.NoError(t, err)
	assert.Equal(t, domain.ContributionStatusApproved, contrib.Status)
	assert.Equal(t, &reviewer, contrib.ReviewerID)
	assert.NotNil(t, contrib.ReviewedAt)
}

func TestContribution_ApproveAlreadyApproved(t *testing.T) {
	contrib := pendingContribution(t)
	_ = contrib.Approve(uuid.New(), "")
	err := contrib.Approve(uuid.New(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestContribution_RejectPending(t *testing.T) {
	contrib := pendingContribution(t)
	reviewer := uuid.New()
	err := contrib.Reject(reviewer, "not accurate")
	require.NoError(t, err)
	assert.Equal(t, domain.ContributionStatusRejected, contrib.Status)
	assert.NotNil(t, contrib.ReviewedAt)
}

func TestContribution_WithdrawByNonOwner(t *testing.T) {
	contrib := pendingContribution(t)
	err := contrib.Withdraw(uuid.New()) // different user
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestContribution_WithdrawAfterApproval(t *testing.T) {
	contrib := pendingContribution(t)
	_ = contrib.Approve(uuid.New(), "")
	err := contrib.Withdraw(contrib.ContributorID)
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestContribution_NewDefinition_RequiresTargetWordID(t *testing.T) {
	_, err := domain.NewContribution(uuid.New(), domain.ContributionTypeNewDefinition, nil, map[string]any{})
	assert.ErrorIs(t, err, domain.ErrTargetRequired)
}

func TestContribution_NewWord_NoTargetRequired(t *testing.T) {
	c, err := domain.NewContribution(uuid.New(), domain.ContributionTypeNewWord, nil, map[string]any{"banjar": "test"})
	require.NoError(t, err)
	assert.Equal(t, domain.ContributionStatusPending, c.Status)
}

func pendingContribution(t *testing.T) *domain.Contribution {
	t.Helper()
	wordID := uuid.New()
	c, err := domain.NewContribution(uuid.New(), domain.ContributionTypeNewDefinition, &wordID, map[string]any{"meaning": "test"})
	require.NoError(t, err)
	return c
}
