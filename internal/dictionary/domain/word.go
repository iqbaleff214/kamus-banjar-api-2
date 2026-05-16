package domain

import (
	"time"

	"github.com/google/uuid"
)

type Word struct {
	ID               uuid.UUID
	Banjar           string
	BanjarSyllabified string
	Dialect          Dialect
	WordClass        WordClass
	HomonymNumber    int
	IsRoot           bool
	RootWordID       *uuid.UUID
	Definitions      []*Definition
	Examples         []*Example
	RelatedWords     []uuid.UUID
	Status           WordStatus
	Source           Source
	SourceReference  string
	CreatedBy        *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

// NewWord creates and validates a new Word aggregate.
func NewWord(banjar string, wordClass WordClass, dialect Dialect) (*Word, error) {
	if banjar == "" {
		return nil, ErrEmptyBanjar
	}
	if err := wordClass.Validate(); err != nil {
		return nil, err
	}
	if err := dialect.Validate(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Word{
		ID:            uuid.New(),
		Banjar:        banjar,
		Dialect:       dialect,
		WordClass:     wordClass,
		HomonymNumber: 1,
		IsRoot:        true,
		Definitions:   make([]*Definition, 0),
		Examples:      make([]*Example, 0),
		RelatedWords:  make([]uuid.UUID, 0),
		Status:        WordStatusActive,
		Source:        SourceSeeded,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// AddDefinition appends a new Definition to the word.
func (w *Word) AddDefinition(meaning string, sortOrder int) error {
	def, err := NewDefinition(meaning, sortOrder, w.Source)
	if err != nil {
		return err
	}
	w.Definitions = append(w.Definitions, def)
	w.UpdatedAt = time.Now().UTC()
	return nil
}

// AddExample appends a new Example to the word.
func (w *Word) AddExample(banjar, indonesian string) {
	ex := NewExample(banjar, indonesian, w.Source)
	w.Examples = append(w.Examples, ex)
	w.UpdatedAt = time.Now().UTC()
}

// Deprecate marks the word as deprecated.
func (w *Word) Deprecate() {
	w.Status = WordStatusDeprecated
	w.UpdatedAt = time.Now().UTC()
}

// Restore sets the word back to active.
func (w *Word) Restore() {
	w.Status = WordStatusActive
	w.UpdatedAt = time.Now().UTC()
}

// SoftDelete sets DeletedAt to mark the word as deleted.
func (w *Word) SoftDelete() {
	now := time.Now().UTC()
	w.DeletedAt = &now
	w.UpdatedAt = now
}

// IsDeleted returns true if the word has been soft-deleted.
func (w *Word) IsDeleted() bool {
	return w.DeletedAt != nil
}
