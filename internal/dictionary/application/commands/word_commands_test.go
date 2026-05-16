package commands_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
)

// ─── in-memory fake ───────────────────────────────────────────────────────────

type fakeWordRepo struct {
	words map[string]*domain.Word
}

func newFakeWordRepo() *fakeWordRepo {
	return &fakeWordRepo{words: make(map[string]*domain.Word)}
}

func (r *fakeWordRepo) Create(_ context.Context, w *domain.Word) error {
	for _, existing := range r.words {
		if existing.Banjar == w.Banjar && existing.WordClass == w.WordClass && existing.Dialect == w.Dialect {
			return domain.ErrWordConflict
		}
	}
	r.words[w.ID.String()] = w
	return nil
}
func (r *fakeWordRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Word, error) {
	w, ok := r.words[id.String()]
	if !ok {
		return nil, domain.ErrWordNotFound
	}
	return w, nil
}
func (r *fakeWordRepo) FindAll(_ context.Context, _ domain.WordFilter, _, _ int) ([]*domain.Word, int, error) {
	return nil, 0, nil
}
func (r *fakeWordRepo) Search(_ context.Context, _ string, _ domain.WordFilter, _, _ int) ([]*domain.Word, int, error) {
	return nil, 0, nil
}
func (r *fakeWordRepo) Update(_ context.Context, w *domain.Word) error {
	if _, ok := r.words[w.ID.String()]; !ok {
		return domain.ErrWordNotFound
	}
	r.words[w.ID.String()] = w
	return nil
}
func (r *fakeWordRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	w, ok := r.words[id.String()]
	if !ok {
		return domain.ErrWordNotFound
	}
	w.SoftDelete()
	return nil
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestCreateWord_Success(t *testing.T) {
	svc := commands.NewWordCommandService(newFakeWordRepo())
	input := commands.WordInput{
		Banjar:      "abah",
		WordClass:   "n",
		Dialect:     "hulu",
		IsRoot:      true,
		Definitions: []commands.DefinitionInput{{Meaning: "ayah"}},
	}
	w, err := svc.CreateWord(context.Background(), uuid.New().String(), input)
	require.NoError(t, err)
	assert.Equal(t, "abah", w.Banjar)
	assert.Equal(t, domain.WordClassNomina, w.WordClass)
	assert.Len(t, w.Definitions, 1)
}

func TestCreateWord_DuplicateConflict(t *testing.T) {
	repo := newFakeWordRepo()
	svc := commands.NewWordCommandService(repo)
	input := commands.WordInput{Banjar: "abah", WordClass: "n", Dialect: "hulu", IsRoot: true}
	_, err := svc.CreateWord(context.Background(), uuid.New().String(), input)
	require.NoError(t, err)

	_, err = svc.CreateWord(context.Background(), uuid.New().String(), input)
	assert.ErrorIs(t, err, domain.ErrWordConflict)
}

func TestCreateWord_InvalidWordClass(t *testing.T) {
	svc := commands.NewWordCommandService(newFakeWordRepo())
	input := commands.WordInput{Banjar: "abah", WordClass: "invalid", Dialect: "hulu"}
	_, err := svc.CreateWord(context.Background(), "", input)
	assert.ErrorIs(t, err, domain.ErrInvalidWordClass)
}

func TestUpdateWord_Success(t *testing.T) {
	repo := newFakeWordRepo()
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	_ = repo.Create(context.Background(), w)

	svc := commands.NewWordCommandService(repo)
	input := commands.WordInput{
		Banjar:      "abah",
		WordClass:   "v",
		Definitions: []commands.DefinitionInput{{Meaning: "updated meaning"}},
	}
	updated, err := svc.UpdateWord(context.Background(), uuid.New().String(), w.ID.String(), input)
	require.NoError(t, err)
	assert.Equal(t, domain.WordClassVerba, updated.WordClass)
	assert.Len(t, updated.Definitions, 1)
	assert.Equal(t, "updated meaning", updated.Definitions[0].Meaning)
}

func TestUpdateWord_NotFound(t *testing.T) {
	svc := commands.NewWordCommandService(newFakeWordRepo())
	_, err := svc.UpdateWord(context.Background(), "", uuid.New().String(), commands.WordInput{})
	assert.ErrorIs(t, err, domain.ErrWordNotFound)
}

func TestUpdateWord_InvalidWordID(t *testing.T) {
	svc := commands.NewWordCommandService(newFakeWordRepo())
	_, err := svc.UpdateWord(context.Background(), "", "not-a-uuid", commands.WordInput{})
	assert.ErrorIs(t, err, domain.ErrWordNotFound)
}

func TestSoftDeleteWord_Success(t *testing.T) {
	repo := newFakeWordRepo()
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	_ = repo.Create(context.Background(), w)

	svc := commands.NewWordCommandService(repo)
	err := svc.SoftDeleteWord(context.Background(), w.ID.String())
	require.NoError(t, err)
	assert.True(t, w.IsDeleted())
}

func TestSoftDeleteWord_AlreadyDeleted(t *testing.T) {
	repo := newFakeWordRepo()
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	_ = repo.Create(context.Background(), w)
	w.SoftDelete()

	svc := commands.NewWordCommandService(repo)
	err := svc.SoftDeleteWord(context.Background(), w.ID.String())
	assert.NoError(t, err) // idempotent
}

func TestSoftDeleteWord_NotFound(t *testing.T) {
	svc := commands.NewWordCommandService(newFakeWordRepo())
	err := svc.SoftDeleteWord(context.Background(), uuid.New().String())
	assert.NoError(t, err) // idempotent for missing too
}
