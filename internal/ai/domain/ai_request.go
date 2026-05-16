package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AIRequestType string

const (
	AIRequestTypeEnrichDefinition AIRequestType = "enrich_definition"
	AIRequestTypeSuggestExample   AIRequestType = "suggest_example"
	AIRequestTypeSuggestRelated   AIRequestType = "suggest_related"
	AIRequestTypeQualityCheck     AIRequestType = "quality_check"
)

type AIRequestStatus string

const (
	AIRequestStatusPending   AIRequestStatus = "pending"
	AIRequestStatusCompleted AIRequestStatus = "completed"
	AIRequestStatusFailed    AIRequestStatus = "failed"
)

type AIReviewStatus string

const (
	AIReviewStatusUnreviewed AIReviewStatus = "unreviewed"
	AIReviewStatusApproved   AIReviewStatus = "approved"
	AIReviewStatusRejected   AIReviewStatus = "rejected"
)

var (
	ErrCannotApproveQualityCheck = errors.New("quality_check requests cannot be approved")
	ErrAIInvalidTransition       = errors.New("invalid state transition")
)

type AIRequest struct {
	ID                   uuid.UUID
	Type                 AIRequestType
	TargetWordID         *uuid.UUID
	TargetContributionID *uuid.UUID
	RequestedBy          uuid.UUID
	Model                string
	Prompt               string
	Response             map[string]any
	ParsedOutput         map[string]any
	Status               AIRequestStatus
	ReviewStatus         AIReviewStatus
	ReviewedBy           *uuid.UUID
	ReviewedAt           *time.Time
	CreatedAt            time.Time
}

func NewAIRequest(
	ctype AIRequestType,
	targetWordID, targetContributionID *uuid.UUID,
	requestedBy uuid.UUID,
	model, prompt string,
) *AIRequest {
	return &AIRequest{
		ID:                   uuid.New(),
		Type:                 ctype,
		TargetWordID:         targetWordID,
		TargetContributionID: targetContributionID,
		RequestedBy:          requestedBy,
		Model:                model,
		Prompt:               prompt,
		Status:               AIRequestStatusPending,
		ReviewStatus:         AIReviewStatusUnreviewed,
		CreatedAt:            time.Now().UTC(),
	}
}

func (r *AIRequest) MarkCompleted(response, parsedOutput map[string]any) error {
	if r.Status != AIRequestStatusPending {
		return ErrAIInvalidTransition
	}
	r.Status = AIRequestStatusCompleted
	r.Response = response
	r.ParsedOutput = parsedOutput
	return nil
}

func (r *AIRequest) MarkFailed(rawError map[string]any) error {
	r.Status = AIRequestStatusFailed
	r.Response = rawError
	return nil
}

func (r *AIRequest) Approve(reviewerID uuid.UUID) error {
	if r.Type == AIRequestTypeQualityCheck {
		return ErrCannotApproveQualityCheck
	}
	if r.ReviewStatus == AIReviewStatusApproved {
		return ErrAIInvalidTransition
	}
	r.ReviewStatus = AIReviewStatusApproved
	r.ReviewedBy = &reviewerID
	now := time.Now().UTC()
	r.ReviewedAt = &now
	return nil
}

func (r *AIRequest) Reject(reviewerID uuid.UUID) error {
	if r.ReviewStatus == AIReviewStatusRejected {
		return ErrAIInvalidTransition
	}
	r.ReviewStatus = AIReviewStatusRejected
	r.ReviewedBy = &reviewerID
	now := time.Now().UTC()
	r.ReviewedAt = &now
	return nil
}
