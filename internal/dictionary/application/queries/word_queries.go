package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/pagination"
)

type WordQueryService struct {
	words domain.WordRepository
}

func NewWordQueryService(words domain.WordRepository) *WordQueryService {
	return &WordQueryService{words: words}
}

type ListResult struct {
	Words []*domain.Word
	Meta  httperr.PaginationMeta
}

// GetWord returns an active, non-deleted word by ID.
func (s *WordQueryService) GetWord(ctx context.Context, idStr string) (*domain.Word, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, domain.ErrWordNotFound
	}
	word, err := s.words.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if word.IsDeleted() || word.Status == domain.WordStatusDeprecated {
		return nil, domain.ErrWordNotFound
	}
	return word, nil
}

// ListWords returns a paginated list of active words, applying filters.
func (s *WordQueryService) ListWords(ctx context.Context, filter domain.WordFilter, page, perPage int) (*ListResult, error) {
	if page < 1 {
		page = pagination.DefaultPage
	}
	if perPage < 1 || perPage > pagination.MaxPerPage {
		perPage = pagination.DefaultPerPage
	}

	activeStatus := domain.WordStatusActive
	if filter.Status == nil {
		filter.Status = &activeStatus
	}

	words, total, err := s.words.FindAll(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Words: words,
		Meta: httperr.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: pagination.TotalPages(total, perPage),
		},
	}, nil
}

// SearchWords performs full-text search. Falls back to ListWords when query is empty.
func (s *WordQueryService) SearchWords(ctx context.Context, query string, filter domain.WordFilter, page, perPage int) (*ListResult, error) {
	if query == "" {
		return s.ListWords(ctx, filter, page, perPage)
	}
	if page < 1 {
		page = pagination.DefaultPage
	}
	if perPage < 1 || perPage > pagination.MaxPerPage {
		perPage = pagination.DefaultPerPage
	}

	activeStatus := domain.WordStatusActive
	if filter.Status == nil {
		filter.Status = &activeStatus
	}

	words, total, err := s.words.Search(ctx, query, filter, page, perPage)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Words: words,
		Meta: httperr.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: pagination.TotalPages(total, perPage),
		},
	}, nil
}
