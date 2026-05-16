package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
)

type WordInput struct {
	Banjar            string
	BanjarSyllabified string
	WordClass         string
	Dialect           string
	HomonymNumber     int
	IsRoot            bool
	RootWordID        *uuid.UUID
	Definitions       []DefinitionInput
	Examples          []ExampleInput
	SourceReference   string
}

type DefinitionInput struct {
	Meaning   string
	SortOrder int
}

type ExampleInput struct {
	BanjarSentence        string
	IndonesianTranslation string
}

type WordCommandService struct {
	words domain.WordRepository
}

func NewWordCommandService(words domain.WordRepository) *WordCommandService {
	return &WordCommandService{words: words}
}

// CreateWord creates a new word directly (admin path, bypasses contribution flow).
func (s *WordCommandService) CreateWord(ctx context.Context, adminID string, input WordInput) (*domain.Word, error) {
	wc := domain.WordClass(input.WordClass)
	dialect := domain.Dialect(input.Dialect)
	if dialect == "" {
		dialect = domain.DialectHulu
	}

	word, err := domain.NewWord(input.Banjar, wc, dialect)
	if err != nil {
		return nil, err
	}
	word.BanjarSyllabified = input.BanjarSyllabified
	word.SourceReference = input.SourceReference
	word.Source = domain.SourceSeeded
	if adminID != "" {
		if id, parseErr := uuid.Parse(adminID); parseErr == nil {
			word.CreatedBy = &id
			word.Source = domain.SourceContributed
		}
	}
	if input.HomonymNumber > 0 {
		word.HomonymNumber = input.HomonymNumber
	}
	word.IsRoot = input.IsRoot
	if !word.IsRoot {
		word.RootWordID = input.RootWordID
	}

	for i, d := range input.Definitions {
		so := d.SortOrder
		if so == 0 {
			so = i + 1
		}
		if err := word.AddDefinition(d.Meaning, so); err != nil {
			return nil, err
		}
	}
	for _, e := range input.Examples {
		word.AddExample(e.BanjarSentence, e.IndonesianTranslation)
	}

	if err := s.words.Create(ctx, word); err != nil {
		return nil, err
	}
	return word, nil
}

// UpdateWord replaces word data (admin path).
func (s *WordCommandService) UpdateWord(ctx context.Context, adminID, wordID string, input WordInput) (*domain.Word, error) {
	id, err := uuid.Parse(wordID)
	if err != nil {
		return nil, domain.ErrWordNotFound
	}
	word, err := s.words.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if word.IsDeleted() {
		return nil, domain.ErrWordNotFound
	}

	if input.Banjar != "" {
		word.Banjar = input.Banjar
	}
	if input.BanjarSyllabified != "" {
		word.BanjarSyllabified = input.BanjarSyllabified
	}
	if input.WordClass != "" {
		wc := domain.WordClass(input.WordClass)
		if err := wc.Validate(); err != nil {
			return nil, err
		}
		word.WordClass = wc
	}
	if input.SourceReference != "" {
		word.SourceReference = input.SourceReference
	}

	word.Definitions = make([]*domain.Definition, 0, len(input.Definitions))
	for i, d := range input.Definitions {
		so := d.SortOrder
		if so == 0 {
			so = i + 1
		}
		if err := word.AddDefinition(d.Meaning, so); err != nil {
			return nil, err
		}
	}
	word.Examples = make([]*domain.Example, 0, len(input.Examples))
	for _, e := range input.Examples {
		word.AddExample(e.BanjarSentence, e.IndonesianTranslation)
	}
	word.UpdatedAt = time.Now().UTC()

	if err := s.words.Update(ctx, word); err != nil {
		return nil, err
	}
	return word, nil
}

// SoftDeleteWord marks a word as deleted.
func (s *WordCommandService) SoftDeleteWord(ctx context.Context, wordID string) error {
	id, err := uuid.Parse(wordID)
	if err != nil {
		return domain.ErrWordNotFound
	}
	word, err := s.words.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWordNotFound) {
			return nil // idempotent
		}
		return err
	}
	if word.IsDeleted() {
		return nil // already deleted — idempotent
	}
	return s.words.SoftDelete(ctx, id)
}
