package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

func TestNewAuditLog_ValidAction(t *testing.T) {
	actorID := uuid.New()
	targetID := uuid.New()
	log, err := domain.NewAuditLog(actorID, domain.ActionBanUser, "user", targetID, nil)
	require.NoError(t, err)
	assert.Equal(t, actorID, log.ActorID)
	assert.Equal(t, domain.ActionBanUser, log.Action)
	assert.Equal(t, "user", log.TargetType)
	assert.Equal(t, targetID, log.TargetID)
	assert.NotEqual(t, uuid.Nil, log.ID)
}

func TestNewAuditLog_InvalidAction(t *testing.T) {
	_, err := domain.NewAuditLog(uuid.New(), domain.AuditAction("not_valid"), "user", uuid.New(), nil)
	assert.ErrorIs(t, err, domain.ErrInvalidAuditAction)
}

func TestNewAuditLog_AllValidActions(t *testing.T) {
	actions := []domain.AuditAction{
		domain.ActionApproveContribution,
		domain.ActionRejectContribution,
		domain.ActionBanUser,
		domain.ActionUnbanUser,
		domain.ActionChangeRole,
		domain.ActionDeleteWord,
		domain.ActionApproveAI,
		domain.ActionRejectAI,
	}
	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			_, err := domain.NewAuditLog(uuid.New(), action, "target", uuid.New(), nil)
			assert.NoError(t, err)
		})
	}
}

func TestNewAuditLog_MetadataPropagated(t *testing.T) {
	meta := map[string]any{"role": "admin"}
	log, err := domain.NewAuditLog(uuid.New(), domain.ActionChangeRole, "user", uuid.New(), meta)
	require.NoError(t, err)
	assert.Equal(t, "admin", log.Metadata["role"])
}
