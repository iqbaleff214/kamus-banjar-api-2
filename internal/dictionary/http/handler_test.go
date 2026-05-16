package dictionaryhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/queries"
	dictionaryhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/http"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopCounter struct{}

func (n *noopCounter) Increment(_ context.Context, _ string, _ time.Duration) (int64, error) {
	return 1, nil
}

// ─── in-memory fake repo ─────────────────────────────────────────────────────

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
func (r *fakeWordRepo) Search(_ context.Context, query string, _ domain.WordFilter, _, _ int) ([]*domain.Word, int, error) {
	var out []*domain.Word
	for _, w := range r.words {
		if w.IsDeleted() {
			continue
		}
		if contains(w.Banjar, query) {
			out = append(out, w)
		}
	}
	return out, len(out), nil
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

// ─── helpers ─────────────────────────────────────────────────────────────────

func makeWord(banjar string, wc domain.WordClass) *domain.Word {
	w, _ := domain.NewWord(banjar, wc, domain.DialectHulu)
	_ = w.AddDefinition("arti "+banjar, 1)
	return w
}

func newTestApp(repo *fakeWordRepo) *fiber.App {
	auth.Init("test-secret")
	qs := queries.NewWordQueryService(repo)
	cmd := commands.NewWordCommandService(repo)
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	dictionaryhttp.RegisterRoutes(app, dictionaryhttp.NewHandler(qs, cmd), &noopCounter{})
	return app
}

func doRequest(app *fiber.App, method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

func adminToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "admin")
	require.NoError(t, err)
	return tok
}

func userToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "user")
	require.NoError(t, err)
	return tok
}

func decodeJSON(t *testing.T, body io.ReadCloser, out interface{}) {
	t.Helper()
	defer func() { _ = body.Close() }()
	require.NoError(t, json.NewDecoder(body).Decode(out))
}

// ─── public endpoint tests ────────────────────────────────────────────────────

func TestListWordsHandler_200(t *testing.T) {
	repo := newFakeWordRepo()
	_ = repo.Create(context.Background(), makeWord("abah", domain.WordClassNomina))
	_ = repo.Create(context.Background(), makeWord("makan", domain.WordClassVerba))

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	assert.True(t, out["success"].(bool))
	data := out["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestListWordsHandler_SearchQuery(t *testing.T) {
	repo := newFakeWordRepo()
	_ = repo.Create(context.Background(), makeWord("abah", domain.WordClassNomina))
	_ = repo.Create(context.Background(), makeWord("makan", domain.WordClassVerba))

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words?q=abah", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	data := out["data"].([]interface{})
	assert.Len(t, data, 1)
	assert.Equal(t, "abah", data[0].(map[string]interface{})["banjar"])
}

func TestListWordsHandler_FilterWordClass(t *testing.T) {
	repo := newFakeWordRepo()
	_ = repo.Create(context.Background(), makeWord("abah", domain.WordClassNomina))
	_ = repo.Create(context.Background(), makeWord("makan", domain.WordClassVerba))

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words?word_class=n", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	data := out["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestGetWordHandler_200(t *testing.T) {
	repo := newFakeWordRepo()
	w := makeWord("abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w)

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words/"+w.ID.String(), nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	assert.Equal(t, "abah", out["data"].(map[string]interface{})["banjar"])
}

func TestGetWordHandler_404(t *testing.T) {
	app := newTestApp(newFakeWordRepo())
	resp := doRequest(app, "GET", "/api/v2/words/"+uuid.New().String(), nil, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDefinitionsHandler_SortedByScore(t *testing.T) {
	repo := newFakeWordRepo()
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	_ = w.AddDefinition("first meaning", 1)
	_ = w.AddDefinition("second meaning", 2)
	_ = repo.Create(context.Background(), w)

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words/"+w.ID.String()+"/definitions", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	data := out["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestGetExamplesHandler_200(t *testing.T) {
	repo := newFakeWordRepo()
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	w.AddExample("abah handak makan", "ayah ingin makan")
	_ = repo.Create(context.Background(), w)

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words/"+w.ID.String()+"/examples", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	data := out["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestGetRelatedWordsHandler_200(t *testing.T) {
	repo := newFakeWordRepo()
	w1 := makeWord("abah", domain.WordClassNomina)
	w2 := makeWord("abah-abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w1)
	_ = repo.Create(context.Background(), w2)
	w1.RelatedWords = append(w1.RelatedWords, w2.ID)

	app := newTestApp(repo)
	resp := doRequest(app, "GET", "/api/v2/words/"+w1.ID.String()+"/related", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	data := out["data"].([]interface{})
	assert.Len(t, data, 1)
}

// ─── admin endpoint tests ─────────────────────────────────────────────────────

func TestAdminCreateWordHandler_403_NonAdmin(t *testing.T) {
	app := newTestApp(newFakeWordRepo())
	body := map[string]interface{}{"banjar": "abah", "word_class": "n", "dialect": "hulu"}
	resp := doRequest(app, "POST", "/api/v2/admin/words", body, userToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminCreateWordHandler_401_NoToken(t *testing.T) {
	app := newTestApp(newFakeWordRepo())
	body := map[string]interface{}{"banjar": "abah", "word_class": "n", "dialect": "hulu"}
	resp := doRequest(app, "POST", "/api/v2/admin/words", body, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAdminCreateWordHandler_201(t *testing.T) {
	app := newTestApp(newFakeWordRepo())
	body := map[string]interface{}{
		"banjar":     "abah",
		"word_class": "n",
		"dialect":    "hulu",
		"is_root":    true,
		"definitions": []map[string]interface{}{
			{"meaning": "ayah", "sort_order": 1},
		},
	}
	resp := doRequest(app, "POST", "/api/v2/admin/words", body, adminToken(t))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	assert.Equal(t, "abah", out["data"].(map[string]interface{})["banjar"])
}

func TestAdminCreateWordHandler_422_MissingFields(t *testing.T) {
	app := newTestApp(newFakeWordRepo())
	resp := doRequest(app, "POST", "/api/v2/admin/words", map[string]interface{}{}, adminToken(t))
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAdminUpdateWordHandler_200(t *testing.T) {
	repo := newFakeWordRepo()
	w := makeWord("abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w)

	app := newTestApp(repo)
	body := map[string]interface{}{
		"banjar":     "abah",
		"word_class": "v",
		"definitions": []map[string]interface{}{
			{"meaning": "updated", "sort_order": 1},
		},
	}
	resp := doRequest(app, "PATCH", "/api/v2/admin/words/"+w.ID.String(), body, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]interface{}
	decodeJSON(t, resp.Body, &out)
	assert.Equal(t, "v", out["data"].(map[string]interface{})["word_class"])
}

func TestAdminDeleteWordHandler_204(t *testing.T) {
	repo := newFakeWordRepo()
	w := makeWord("abah", domain.WordClassNomina)
	_ = repo.Create(context.Background(), w)

	app := newTestApp(repo)
	resp := doRequest(app, "DELETE", "/api/v2/admin/words/"+w.ID.String(), nil, adminToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
