package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
)

func newPendingRequest(t domain.AIRequestType) *domain.AIRequest {
	return domain.NewAIRequest(t, nil, nil, uuid.New(), "model", "prompt")
}

func TestAIRequest_MarkCompleted_StateTransition(t *testing.T) {
	r := newPendingRequest(domain.AIRequestTypeEnrichDefinition)
	err := r.MarkCompleted(map[string]any{"raw": "ok"}, map[string]any{"result": "ok"})
	require.NoError(t, err)
	assert.Equal(t, domain.AIRequestStatusCompleted, r.Status)

	err = r.MarkCompleted(map[string]any{}, map[string]any{})
	assert.ErrorIs(t, err, domain.ErrAIInvalidTransition)
}

func TestAIRequest_ApproveQualityCheck_Error(t *testing.T) {
	r := newPendingRequest(domain.AIRequestTypeQualityCheck)
	_ = r.MarkCompleted(map[string]any{}, map[string]any{})

	err := r.Approve(uuid.New())
	assert.ErrorIs(t, err, domain.ErrCannotApproveQualityCheck)
}

func TestAIRequest_ApproveAlreadyApproved_Error(t *testing.T) {
	r := newPendingRequest(domain.AIRequestTypeEnrichDefinition)
	_ = r.MarkCompleted(map[string]any{}, map[string]any{})

	reviewer := uuid.New()
	require.NoError(t, r.Approve(reviewer))

	err := r.Approve(reviewer)
	assert.ErrorIs(t, err, domain.ErrAIInvalidTransition)
}

func TestAIRequest_RejectAlreadyRejected_Error(t *testing.T) {
	r := newPendingRequest(domain.AIRequestTypeEnrichDefinition)
	_ = r.MarkCompleted(map[string]any{}, map[string]any{})

	reviewer := uuid.New()
	require.NoError(t, r.Reject(reviewer))

	err := r.Reject(reviewer)
	assert.ErrorIs(t, err, domain.ErrAIInvalidTransition)
}

func TestAIRequest_MarkFailed(t *testing.T) {
	r := newPendingRequest(domain.AIRequestTypeSuggestExample)
	err := r.MarkFailed(map[string]any{"error": "timeout"})
	require.NoError(t, err)
	assert.Equal(t, domain.AIRequestStatusFailed, r.Status)
}
