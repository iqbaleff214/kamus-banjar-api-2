package domain

import (
	"time"

	"github.com/google/uuid"
)

type Bookmark struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	WordID    uuid.UUID
	CreatedAt time.Time
}

func NewBookmark(userID, wordID uuid.UUID) *Bookmark {
	return &Bookmark{
		ID:        uuid.New(),
		UserID:    userID,
		WordID:    wordID,
		CreatedAt: time.Now().UTC(),
	}
}
