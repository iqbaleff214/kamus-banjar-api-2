package communityhttp

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

// ─── Request bodies ───────────────────────────────────────────────────────────

type submitContributionRequest struct {
	Type         string         `json:"type"`
	TargetWordID *string        `json:"target_word_id"`
	Payload      map[string]any `json:"payload"`
}

type castVoteRequest struct {
	Value string `json:"value"`
}

type addBookmarkRequest struct {
	WordID string `json:"word_id"`
}

type postCommentRequest struct {
	Body string `json:"body"`
}

type editCommentRequest struct {
	Body string `json:"body"`
}

type rejectContributionRequest struct {
	Note string `json:"note"`
}

type approveContributionRequest struct {
	Note string `json:"note"`
}

// ─── Response shapes ──────────────────────────────────────────────────────────

type contributionResponse struct {
	ID            uuid.UUID  `json:"id"`
	Type          string     `json:"type"`
	ContributorID uuid.UUID  `json:"contributor_id"`
	TargetWordID  *uuid.UUID `json:"target_word_id"`
	Payload       any        `json:"payload"`
	Status        string     `json:"status"`
	ReviewerID    *uuid.UUID `json:"reviewer_id,omitempty"`
	ReviewerNote  *string    `json:"reviewer_note,omitempty"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
}

func toContributionResponse(c *domain.Contribution) contributionResponse {
	return contributionResponse{
		ID:            c.ID,
		Type:          string(c.Type),
		ContributorID: c.ContributorID,
		TargetWordID:  c.TargetWordID,
		Payload:       c.Payload,
		Status:        string(c.Status),
		ReviewerID:    c.ReviewerID,
		ReviewerNote:  c.ReviewerNote,
		SubmittedAt:   c.SubmittedAt,
		ReviewedAt:    c.ReviewedAt,
	}
}

type voteResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Value      string    `json:"value"`
	CreatedAt  time.Time `json:"created_at"`
}

func toVoteResponse(v *domain.Vote) voteResponse {
	return voteResponse{
		ID:         v.ID,
		UserID:     v.UserID,
		TargetType: string(v.TargetType),
		TargetID:   v.TargetID,
		Value:      string(v.Value),
		CreatedAt:  v.CreatedAt,
	}
}

type bookmarkResponse struct {
	ID        uuid.UUID `json:"id"`
	WordID    uuid.UUID `json:"word_id"`
	CreatedAt time.Time `json:"created_at"`
}

func toBookmarkResponse(b *domain.Bookmark) bookmarkResponse {
	return bookmarkResponse{
		ID:        b.ID,
		WordID:    b.WordID,
		CreatedAt: b.CreatedAt,
	}
}

type commentResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Body       string    `json:"body"`
	IsFlagged  bool      `json:"is_flagged"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func toCommentResponse(c *domain.Comment) commentResponse {
	return commentResponse{
		ID:         c.ID,
		UserID:     c.UserID,
		TargetType: string(c.TargetType),
		TargetID:   c.TargetID,
		Body:       c.Body,
		IsFlagged:  c.IsFlagged,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}
