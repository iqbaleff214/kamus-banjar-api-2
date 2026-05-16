package commands

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

type CommentService struct {
	comments domain.CommentRepository
	words    ExistChecker
}

func NewCommentService(comments domain.CommentRepository, words ExistChecker) *CommentService {
	return &CommentService{comments: comments, words: words}
}

func (s *CommentService) PostComment(ctx context.Context, userID, wordID uuid.UUID, body string) (*domain.Comment, error) {
	exists, err := s.words.Exists(ctx, wordID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCommentWordNotFound
	}

	c, err := domain.NewComment(userID, domain.CommentTargetWord, wordID, body)
	if err != nil {
		return nil, err
	}
	return s.comments.Create(ctx, c)
}

func (s *CommentService) EditComment(ctx context.Context, callerID, commentID uuid.UUID, body string) (*domain.Comment, error) {
	c, err := s.comments.FindByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if err := c.Edit(callerID, body); err != nil {
		return nil, err
	}
	return s.comments.Update(ctx, c)
}

func (s *CommentService) DeleteComment(ctx context.Context, callerID uuid.UUID, callerRole string, commentID uuid.UUID) error {
	c, err := s.comments.FindByID(ctx, commentID)
	if err != nil {
		return err
	}
	if callerRole != "admin" && c.UserID != callerID {
		return domain.ErrForbidden
	}
	return s.comments.Delete(ctx, commentID)
}

func (s *CommentService) FlagComment(ctx context.Context, commentID uuid.UUID) (*domain.Comment, error) {
	c, err := s.comments.FindByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	c.Flag()
	return s.comments.Update(ctx, c)
}

func (s *CommentService) ListComments(ctx context.Context, wordID uuid.UUID, page, perPage int) ([]*domain.Comment, int, error) {
	exists, err := s.words.Exists(ctx, wordID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, ErrCommentWordNotFound
	}
	return s.comments.FindByTarget(ctx, domain.CommentTargetWord, wordID, page, perPage)
}

var (
	ErrCommentWordNotFound = errors.New("word not found")
	ErrCommentForbidden    = errors.New("action not permitted")
)

// silence unused var
var _ = errors.New
