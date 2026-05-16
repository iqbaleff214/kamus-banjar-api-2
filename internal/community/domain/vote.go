package domain

import (
	"time"

	"github.com/google/uuid"
)

type VoteTargetType string

const (
	VoteTargetWord       VoteTargetType = "word"
	VoteTargetDefinition VoteTargetType = "definition"
)

type VoteValue string

const (
	VoteUp   VoteValue = "up"
	VoteDown VoteValue = "down"
)

type Vote struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TargetType VoteTargetType
	TargetID   uuid.UUID
	Value      VoteValue
	CreatedAt  time.Time
}

func CastVote(userID uuid.UUID, targetType VoteTargetType, targetID uuid.UUID, value VoteValue) (*Vote, error) {
	if targetType != VoteTargetWord && targetType != VoteTargetDefinition {
		return nil, ErrInvalidVoteTarget
	}
	if value != VoteUp && value != VoteDown {
		return nil, ErrInvalidVoteValue
	}
	return &Vote{
		ID:         uuid.New(),
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		Value:      value,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
