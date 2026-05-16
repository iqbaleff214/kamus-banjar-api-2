package commands

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
)

type BookmarkService struct {
	bookmarks domain.BookmarkRepository
	words     ExistChecker
}

func NewBookmarkService(bookmarks domain.BookmarkRepository, words ExistChecker) *BookmarkService {
	return &BookmarkService{bookmarks: bookmarks, words: words}
}

func (s *BookmarkService) AddBookmark(ctx context.Context, userID, wordID uuid.UUID) (*domain.Bookmark, error) {
	exists, err := s.words.Exists(ctx, wordID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrBookmarkWordNotFound
	}

	b := domain.NewBookmark(userID, wordID)
	result, err := s.bookmarks.Create(ctx, b)
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkConflict) {
			return nil, domain.ErrBookmarkConflict
		}
		return nil, err
	}
	return result, nil
}

func (s *BookmarkService) RemoveBookmark(ctx context.Context, userID, wordID uuid.UUID) error {
	exists, err := s.bookmarks.Exists(ctx, userID, wordID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrBookmarkNotFound
	}
	return s.bookmarks.Delete(ctx, userID, wordID)
}

func (s *BookmarkService) ListBookmarks(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*domain.Bookmark, int, error) {
	return s.bookmarks.FindByUser(ctx, userID, page, perPage)
}

var ErrBookmarkWordNotFound = errors.New("word not found")
