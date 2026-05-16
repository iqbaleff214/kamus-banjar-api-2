package domain

import (
	"time"

	"github.com/google/uuid"
)

type ContributionType string

const (
	ContributionTypeNewWord       ContributionType = "new_word"
	ContributionTypeNewDefinition ContributionType = "new_definition"
	ContributionTypeNewExample    ContributionType = "new_example"
	ContributionTypeEditWord      ContributionType = "edit_word"
)

type ContributionStatus string

const (
	ContributionStatusPending   ContributionStatus = "pending"
	ContributionStatusApproved  ContributionStatus = "approved"
	ContributionStatusRejected  ContributionStatus = "rejected"
	ContributionStatusWithdrawn ContributionStatus = "withdrawn"
)

type Contribution struct {
	ID            uuid.UUID
	Type          ContributionType
	ContributorID uuid.UUID
	TargetWordID  *uuid.UUID
	Payload       map[string]any
	Status        ContributionStatus
	ReviewerID    *uuid.UUID
	ReviewerNote  *string
	SubmittedAt   time.Time
	ReviewedAt    *time.Time
}

func NewContribution(contributorID uuid.UUID, ctype ContributionType, targetWordID *uuid.UUID, payload map[string]any) (*Contribution, error) {
	if ctype != ContributionTypeNewWord && targetWordID == nil {
		return nil, ErrTargetRequired
	}
	return &Contribution{
		ID:            uuid.New(),
		Type:          ctype,
		ContributorID: contributorID,
		TargetWordID:  targetWordID,
		Payload:       payload,
		Status:        ContributionStatusPending,
		SubmittedAt:   time.Now().UTC(),
	}, nil
}

func (c *Contribution) Approve(reviewerID uuid.UUID, note string) error {
	if c.Status != ContributionStatusPending {
		return ErrInvalidTransition
	}
	c.Status = ContributionStatusApproved
	c.ReviewerID = &reviewerID
	if note != "" {
		c.ReviewerNote = &note
	}
	now := time.Now().UTC()
	c.ReviewedAt = &now
	return nil
}

func (c *Contribution) Reject(reviewerID uuid.UUID, note string) error {
	if c.Status != ContributionStatusPending {
		return ErrInvalidTransition
	}
	c.Status = ContributionStatusRejected
	c.ReviewerID = &reviewerID
	if note != "" {
		c.ReviewerNote = &note
	}
	now := time.Now().UTC()
	c.ReviewedAt = &now
	return nil
}

func (c *Contribution) Withdraw(callerID uuid.UUID) error {
	if c.ContributorID != callerID {
		return ErrForbidden
	}
	if c.Status != ContributionStatusPending {
		return ErrInvalidTransition
	}
	c.Status = ContributionStatusWithdrawn
	return nil
}
