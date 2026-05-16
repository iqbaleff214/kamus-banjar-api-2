package queries_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/queries"
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
func (r *fakeWordRepo) FindAll(_ context.Context, filter domain.WordFilter, page, perPage int) ([]*domain.Word, int, error) {
	var out []*domain.Word
	for _, w := range r.words {
		if w.IsDeleted() {
			continue
		}
		if filter.WordClass != nil && w.WordClass != *filter.WordClass {
			continue
		}
		if filter.IsRoot != nil && w.IsRoot != *filter.IsRoot {
			continue
		}
		if filter.Status != nil && w.Status != *filter.Status {
			continue
		}
		out = append(out, w)
	}
	total := len(out)
	start := (page - 1) * perPage
	if start >= total {
		return []*domain.Word{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return out[start:end], total, nil
}
func (r *fakeWordRepo) Search(_ context.Context, query string, filter domain.WordFilter, page, perPage int) ([]*domain.Word, int, error) {
	var out []*domain.Word
	for _, w := range r.words {
		if w.IsDeleted() {
			continue
		}
		matched := false
		if contains(w.Banjar, query) {
			matched = true
		}
		for _, d := range w.Definitions {
			if contains(d.Meaning, query) {
				matched = true
			}
		}
		if !matched {
			continue
		}
		if filter.WordClass != nil && w.WordClass != *filter.WordClass {
			continue
		}
		out = append(out, w)
	}
	total := len(out)
	return out, total, nil
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

func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func makeWord(banjar string, wc domain.WordClass) *domain.Word {
	w, _ := domain.NewWord(banjar, wc, domain.DialectHulu)
	_ = w.AddDefinition("arti dari "+banjar, 1)
	return w
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestGetWord_Found(t *testing.T) {
	repo := newFakeWordRepo()
	w := makeWord("abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w)

	svc := queries.NewWordQueryService(repo)
	got, err := svc.GetWord(context.Background(), w.ID.String())
	require.NoError(t, err)
	assert.Equal(t, "abah", got.Banjar)
}

func TestGetWord_NotFound(t *testing.T) {
	svc := queries.NewWordQueryService(newFakeWordRepo())
	_, err := svc.GetWord(context.Background(), uuid.New().String())
	assert.ErrorIs(t, err, domain.ErrWordNotFound)
}

func TestGetWord_SoftDeleted(t *testing.T) {
	repo := newFakeWordRepo()
	w := makeWord("abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w)
	w.SoftDelete()

	svc := queries.NewWordQueryService(repo)
	_, err := svc.GetWord(context.Background(), w.ID.String())
	assert.ErrorIs(t, err, domain.ErrWordNotFound)
}

func TestListWords_Pagination(t *testing.T) {
	repo := newFakeWordRepo()
	for i := 0; i < 5; i++ {
		w, _ := domain.NewWord("word"+string(rune('a'+i)), domain.WordClassNomina, domain.DialectHulu)
		_ = repo.Create(context.Background(), w)
	}
	svc := queries.NewWordQueryService(repo)
	result, err := svc.ListWords(context.Background(), domain.WordFilter{}, 1, 3)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(result.Words), 3)
	assert.Equal(t, 5, result.Meta.Total)
	assert.Equal(t, 2, result.Meta.TotalPages)
}

func TestListWords_FilterWordClass(t *testing.T) {
	repo := newFakeWordRepo()
	w1, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	w2, _ := domain.NewWord("makan", domain.WordClassVerba, domain.DialectHulu)
	_ = repo.Create(context.Background(), w1)
	_ = repo.Create(context.Background(), w2)

	svc := queries.NewWordQueryService(repo)
	wc := domain.WordClassNomina
	result, err := svc.ListWords(context.Background(), domain.WordFilter{WordClass: &wc}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Meta.Total)
	assert.Equal(t, "abah", result.Words[0].Banjar)
}

func TestSearchWords_ReturnsRelevantResults(t *testing.T) {
	repo := newFakeWordRepo()
	w1 := makeWord("abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w1)

	svc := queries.NewWordQueryService(repo)
	result, err := svc.SearchWords(context.Background(), "abah", domain.WordFilter{}, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Meta.Total, 1)
}

func TestSearchWords_EmptyQuery_FallsbackToList(t *testing.T) {
	repo := newFakeWordRepo()
	for i := 0; i < 3; i++ {
		w, _ := domain.NewWord("word"+string(rune('a'+i)), domain.WordClassNomina, domain.DialectHulu)
		_ = repo.Create(context.Background(), w)
	}
	svc := queries.NewWordQueryService(repo)
	result, err := svc.SearchWords(context.Background(), "", domain.WordFilter{}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 3, result.Meta.Total)
}
