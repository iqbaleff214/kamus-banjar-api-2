package domain

import (
	"time"

	"github.com/google/uuid"
)

type CommentTargetType string

const (
	CommentTargetWord         CommentTargetType = "word"
	CommentTargetContribution CommentTargetType = "contribution"
)

const maxCommentBody = 1000

type Comment struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TargetType CommentTargetType
	TargetID   uuid.UUID
	Body       string
	IsFlagged  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewComment(userID uuid.UUID, targetType CommentTargetType, targetID uuid.UUID, body string) (*Comment, error) {
	if body == "" {
		return nil, ErrEmptyBody
	}
	if len([]rune(body)) > maxCommentBody {
		return nil, ErrBodyTooLong
	}
	now := time.Now().UTC()
	return &Comment{
		ID:         uuid.New(),
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		Body:       body,
		IsFlagged:  false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (c *Comment) Edit(callerID uuid.UUID, newBody string) error {
	if c.UserID != callerID {
		return ErrForbidden
	}
	if newBody == "" {
		return ErrEmptyBody
	}
	if len([]rune(newBody)) > maxCommentBody {
		return ErrBodyTooLong
	}
	c.Body = newBody
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Comment) Flag() {
	c.IsFlagged = true
	c.UpdatedAt = time.Now().UTC()
}
