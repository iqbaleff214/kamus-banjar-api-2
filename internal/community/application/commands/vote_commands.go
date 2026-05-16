package commands

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

type VoteService struct {
	votes domain.VoteRepository
	words ExistChecker
	defs  ExistChecker
}

func NewVoteService(votes domain.VoteRepository, words, defs ExistChecker) *VoteService {
	return &VoteService{votes: votes, words: words, defs: defs}
}

func (s *VoteService) CastVote(ctx context.Context, userID uuid.UUID, targetType domain.VoteTargetType, targetID uuid.UUID, value domain.VoteValue) (*domain.Vote, error) {
	var checker ExistChecker
	switch targetType {
	case domain.VoteTargetWord:
		checker = s.words
	case domain.VoteTargetDefinition:
		checker = s.defs
	default:
		return nil, domain.ErrInvalidVoteTarget
	}

	exists, err := checker.Exists(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrVoteTargetNotFound
	}

	v, err := domain.CastVote(userID, targetType, targetID, value)
	if err != nil {
		return nil, err
	}

	return s.votes.Upsert(ctx, v)
}

func (s *VoteService) RemoveVote(ctx context.Context, userID uuid.UUID, targetType domain.VoteTargetType, targetID uuid.UUID) error {
	_, err := s.votes.FindByUserAndTarget(ctx, userID, targetType, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrVoteNotFound) {
			return ErrVoteTargetNotFound
		}
		return err
	}
	return s.votes.Delete(ctx, userID, targetType, targetID)
}

var ErrVoteTargetNotFound = errors.New("vote target not found")
