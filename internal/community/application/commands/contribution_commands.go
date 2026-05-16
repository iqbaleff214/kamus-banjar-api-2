package commands

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

// ExistChecker verifies a resource exists by ID.
type ExistChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// UserEmailVerifier checks whether a user's email is verified.
type UserEmailVerifier interface {
	IsEmailVerified(ctx context.Context, userID uuid.UUID) (bool, error)
}

type ContributionInput struct {
	Type         domain.ContributionType
	TargetWordID *uuid.UUID
	Payload      map[string]any
}

type ContributionService struct {
	contributions domain.ContributionRepository
	words         ExistChecker
	userVerifier  UserEmailVerifier
}

func NewContributionService(contributions domain.ContributionRepository, words ExistChecker, userVerifier UserEmailVerifier) *ContributionService {
	return &ContributionService{
		contributions: contributions,
		words:         words,
		userVerifier:  userVerifier,
	}
}

func (s *ContributionService) SubmitContribution(ctx context.Context, userID uuid.UUID, input ContributionInput) (*domain.Contribution, error) {
	verified, err := s.userVerifier.IsEmailVerified(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, ErrEmailNotVerified
	}

	if input.Type != domain.ContributionTypeNewWord && input.TargetWordID != nil {
		exists, err := s.words.Exists(ctx, *input.TargetWordID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrTargetWordNotFound
		}
	}

	c, err := domain.NewContribution(userID, input.Type, input.TargetWordID, input.Payload)
	if err != nil {
		return nil, err
	}

	if err := s.contributions.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ContributionService) WithdrawContribution(ctx context.Context, callerID, contributionID uuid.UUID) (*domain.Contribution, error) {
	c, err := s.contributions.FindByID(ctx, contributionID)
	if err != nil {
		return nil, err
	}

	if err := c.Withdraw(callerID); err != nil {
		return nil, err
	}

	if err := s.contributions.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ContributionService) GetContribution(ctx context.Context, callerID uuid.UUID, callerRole string, contributionID uuid.UUID) (*domain.Contribution, error) {
	c, err := s.contributions.FindByID(ctx, contributionID)
	if err != nil {
		return nil, err
	}
	if callerRole != "admin" && c.ContributorID != callerID {
		return nil, ErrContributionForbidden
	}
	return c, nil
}

func (s *ContributionService) ListContributions(ctx context.Context, callerID uuid.UUID, callerRole string, filter domain.ContributionFilter, page, perPage int) ([]*domain.Contribution, int, error) {
	if callerRole == "admin" {
		return s.contributions.FindAll(ctx, filter, page, perPage)
	}
	return s.contributions.FindByContributor(ctx, callerID, filter, page, perPage)
}

func (s *ContributionService) Update(ctx context.Context, c *domain.Contribution) error {
	return s.contributions.Update(ctx, c)
}

var (
	ErrEmailNotVerified      = errors.New("email must be verified before contributing")
	ErrTargetWordNotFound    = errors.New("target word not found")
	ErrContributionForbidden = errors.New("access to this contribution is not permitted")
)
