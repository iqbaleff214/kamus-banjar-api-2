package moderationhttp

import (
	"time"

	"github.com/google/uuid"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

// ─── Request bodies ───────────────────────────────────────────────────────────

type banUserRequest struct {
	Reason string `json:"reason"`
}

type changeRoleRequest struct {
	Role string `json:"role"`
}

// ─── Response shapes ──────────────────────────────────────────────────────────

type userResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"is_active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func toUserResponse(u *identitydomain.User) userResponse {
	return userResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Role:            string(u.Role),
		IsActive:        u.IsActive,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

type statsResponse struct {
	PendingContributions int `json:"pending_contributions"`
	FlaggedComments      int `json:"flagged_comments"`
	ApprovedThisWeek     int `json:"approved_this_week"`
	RejectedThisWeek     int `json:"rejected_this_week"`
}

func toStatsResponse(s *moderationdomain.ModerationStats) statsResponse {
	return statsResponse{
		PendingContributions: s.PendingContributions,
		FlaggedComments:      s.FlaggedComments,
		ApprovedThisWeek:     s.ApprovedThisWeek,
		RejectedThisWeek:     s.RejectedThisWeek,
	}
}

type contributionResponse struct {
	ID            uuid.UUID  `json:"id"`
	Type          string     `json:"type"`
	ContributorID uuid.UUID  `json:"contributor_id"`
	TargetWordID  *uuid.UUID `json:"target_word_id,omitempty"`
	Payload       any        `json:"payload"`
	Status        string     `json:"status"`
	ReviewerID    *uuid.UUID `json:"reviewer_id,omitempty"`
	ReviewerNote  *string    `json:"reviewer_note,omitempty"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
}

func toContributionResponse(c *communitydomain.Contribution) contributionResponse {
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

func toCommentResponse(c *communitydomain.Comment) commentResponse {
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
