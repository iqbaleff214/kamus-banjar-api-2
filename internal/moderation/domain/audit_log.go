package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	ActionApproveContribution AuditAction = "approve_contribution"
	ActionRejectContribution  AuditAction = "reject_contribution"
	ActionBanUser             AuditAction = "ban_user"
	ActionUnbanUser           AuditAction = "unban_user"
	ActionChangeRole          AuditAction = "change_role"
	ActionDeleteWord          AuditAction = "delete_word"
	ActionApproveAI           AuditAction = "approve_ai"
	ActionRejectAI            AuditAction = "reject_ai"
)

var validActions = map[AuditAction]struct{}{
	ActionApproveContribution: {},
	ActionRejectContribution:  {},
	ActionBanUser:             {},
	ActionUnbanUser:           {},
	ActionChangeRole:          {},
	ActionDeleteWord:          {},
	ActionApproveAI:           {},
	ActionRejectAI:            {},
}

type AuditLog struct {
	ID         uuid.UUID
	ActorID    uuid.UUID
	Action     AuditAction
	TargetType string
	TargetID   uuid.UUID
	Metadata   map[string]any
	CreatedAt  time.Time
}

func NewAuditLog(actorID uuid.UUID, action AuditAction, targetType string, targetID uuid.UUID, metadata map[string]any) (*AuditLog, error) {
	if _, ok := validActions[action]; !ok {
		return nil, ErrInvalidAuditAction
	}
	return &AuditLog{
		ID:         uuid.New(),
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
