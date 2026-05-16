package aihttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	aidomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	aihttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/http"
	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

// ─── fakes ────────────────────────────────────────────────────────────────────

type fakeAIRepo struct {
	store map[uuid.UUID]*aidomain.AIRequest
}

func newFakeAIRepo() *fakeAIRepo {
	return &fakeAIRepo{store: make(map[uuid.UUID]*aidomain.AIRequest)}
}
func (f *fakeAIRepo) Create(_ context.Context, r *aidomain.AIRequest) error {
	f.store[r.ID] = r
	return nil
}
func (f *fakeAIRepo) FindByID(_ context.Context, id uuid.UUID) (*aidomain.AIRequest, error) {
	if r, ok := f.store[id]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeAIRepo) Update(_ context.Context, r *aidomain.AIRequest) error {
	f.store[r.ID] = r
	return nil
}
func (f *fakeAIRepo) ListByWord(_ context.Context, _ uuid.UUID, _, _ int) ([]*aidomain.AIRequest, int, error) {
	var out []*aidomain.AIRequest
	for _, r := range f.store {
		out = append(out, r)
	}
	return out, len(out), nil
}
func (f *fakeAIRepo) ListPendingReview(_ context.Context, _, _ int) ([]*aidomain.AIRequest, int, error) {
	return nil, 0, nil
}

type fakeWordAccessor struct {
	store map[uuid.UUID]*dictdomain.Word
}

func newFakeWordAccessor(words ...*dictdomain.Word) *fakeWordAccessor {
	m := make(map[uuid.UUID]*dictdomain.Word)
	for _, w := range words {
		m[w.ID] = w
	}
	return &fakeWordAccessor{store: m}
}
func (f *fakeWordAccessor) FindByID(_ context.Context, id uuid.UUID) (*dictdomain.Word, error) {
	if w, ok := f.store[id]; ok {
		return w, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeWordAccessor) FindByBanjar(_ context.Context, _ string) (*dictdomain.Word, error) {
	return nil, errors.New("not found")
}
func (f *fakeWordAccessor) Update(_ context.Context, w *dictdomain.Word) error {
	f.store[w.ID] = w
	return nil
}

type fakeContribReader struct {
	store map[uuid.UUID]*communitydomain.Contribution
}

func newFakeContribReader(cs ...*communitydomain.Contribution) *fakeContribReader {
	m := make(map[uuid.UUID]*communitydomain.Contribution)
	for _, c := range cs {
		m[c.ID] = c
	}
	return &fakeContribReader{store: m}
}
func (f *fakeContribReader) FindByID(_ context.Context, id uuid.UUID) (*communitydomain.Contribution, error) {
	if c, ok := f.store[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

type fakeAuditLogWriter struct{}

func (f *fakeAuditLogWriter) CreateAuditLog(_ context.Context, _ *moderationdomain.AuditLog) error {
	return nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func buildAdminApp(t *testing.T, aiRepo *fakeAIRepo, wordRepo *fakeWordAccessor, contribRepo *fakeContribReader, llm aidomain.LLMClient, counter ratelimit.Counter) *fiber.App {
	t.Helper()
	auth.Init("test-secret-that-is-long-enough-32ch")
	enrichSvc := commands.NewEnrichmentService(aiRepo, wordRepo, contribRepo, &fakeAuditLogWriter{}, llm, "test-model")
	translateSvc := commands.NewTranslateService(llm, "test-model")
	app := fiber.New()
	aihttp.RegisterRoutes(app, aihttp.NewHandler(translateSvc), aihttp.NewAdminHandler(enrichSvc), counter)
	return app
}

func adminTok(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "admin")
	require.NoError(t, err)
	return "Bearer " + tok
}

func userTok(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "user")
	require.NoError(t, err)
	return "Bearer " + tok
}

func doAdmin(t *testing.T, app *fiber.App, method, path, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func newTestWord() *dictdomain.Word {
	w, _ := dictdomain.NewWord("urang", dictdomain.WordClassNomina, dictdomain.DialectHulu)
	return w
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestEnrichHandler_202(t *testing.T) {
	word := newTestWord()
	llm := &fakeLLM{response: &aidomain.CompletionResponse{Model: "m", Content: `{"definition":"manusia"}`}}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(word), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/enrich/"+word.ID.String(), adminTok(t))
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestEnrichHandler_404_WordNotFound(t *testing.T) {
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/enrich/"+uuid.New().String(), adminTok(t))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestEnrichHandler_403_NonAdmin(t *testing.T) {
	word := newTestWord()
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(word), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/enrich/"+word.ID.String(), userTok(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestEnrichHandler_429_RateLimit(t *testing.T) {
	word := newTestWord()
	llm := &fakeLLM{response: &aidomain.CompletionResponse{Model: "m", Content: `{}`}}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(word), newFakeContribReader(), llm, nearLimitCounter(50))

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/enrich/"+word.ID.String(), adminTok(t))
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestListAIRequestsHandler_FilterByType(t *testing.T) {
	word := newTestWord()
	aiRepo := newFakeAIRepo()
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, &word.ID, nil, uuid.New(), "m", "p")
	_ = aiRepo.Create(context.Background(), req)

	app := buildAdminApp(t, aiRepo, newFakeWordAccessor(word), newFakeContribReader(), &fakeLLM{}, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodGet, "/api/v2/admin/ai/requests?word_id="+word.ID.String(), adminTok(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestApproveAIRequestHandler_409_QualityCheck(t *testing.T) {
	aiRepo := newFakeAIRepo()
	r := aidomain.NewAIRequest(aidomain.AIRequestTypeQualityCheck, nil, nil, uuid.New(), "m", "p")
	_ = r.MarkCompleted(map[string]any{}, map[string]any{"score": 80})
	_ = aiRepo.Create(context.Background(), r)

	app := buildAdminApp(t, aiRepo, newFakeWordAccessor(), newFakeContribReader(), &fakeLLM{}, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPatch, "/api/v2/admin/ai/requests/"+r.ID.String()+"/approve", adminTok(t))
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestApproveAIRequestHandler_200(t *testing.T) {
	word := newTestWord()
	aiRepo := newFakeAIRepo()
	r := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, &word.ID, nil, uuid.New(), "m", "p")
	_ = r.MarkCompleted(map[string]any{}, map[string]any{"definition": "manusia"})
	_ = aiRepo.Create(context.Background(), r)

	app := buildAdminApp(t, aiRepo, newFakeWordAccessor(word), newFakeContribReader(), &fakeLLM{}, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPatch, "/api/v2/admin/ai/requests/"+r.ID.String()+"/approve", adminTok(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRejectAIRequestHandler_200(t *testing.T) {
	aiRepo := newFakeAIRepo()
	r := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, nil, nil, uuid.New(), "m", "p")
	_ = r.MarkCompleted(map[string]any{}, map[string]any{})
	_ = aiRepo.Create(context.Background(), r)

	app := buildAdminApp(t, aiRepo, newFakeWordAccessor(), newFakeContribReader(), &fakeLLM{}, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPatch, "/api/v2/admin/ai/requests/"+r.ID.String()+"/reject", adminTok(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─── SuggestExample tests ─────────────────────────────────────────────────────

func TestSuggestExampleHandler_202(t *testing.T) {
	word := newTestWord()
	llm := &fakeLLM{response: &aidomain.CompletionResponse{Model: "m", Content: `{"banjar":"inya kada tahu","indonesian":"dia tidak tahu"}`}}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(word), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/example/"+word.ID.String(), adminTok(t))
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestSuggestExampleHandler_404_WordNotFound(t *testing.T) {
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/example/"+uuid.New().String(), adminTok(t))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSuggestExampleHandler_403_NonAdmin(t *testing.T) {
	word := newTestWord()
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(word), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/example/"+word.ID.String(), userTok(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ─── SuggestRelated tests ─────────────────────────────────────────────────────

func TestSuggestRelatedHandler_202(t *testing.T) {
	word := newTestWord()
	llm := &fakeLLM{response: &aidomain.CompletionResponse{Model: "m", Content: `{"related":["urang lain"]}`}}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(word), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/related/"+word.ID.String(), adminTok(t))
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestSuggestRelatedHandler_404_WordNotFound(t *testing.T) {
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/related/"+uuid.New().String(), adminTok(t))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─── QualityCheck tests ───────────────────────────────────────────────────────

func TestQualityCheckHandler_202(t *testing.T) {
	contrib := &communitydomain.Contribution{
		ID:   uuid.New(),
		Type: communitydomain.ContributionTypeNewWord,
	}
	llm := &fakeLLM{response: &aidomain.CompletionResponse{Model: "m", Content: `{"score":85}`}}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(contrib), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/check/"+contrib.ID.String(), adminTok(t))
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestQualityCheckHandler_404_ContribNotFound(t *testing.T) {
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/check/"+uuid.New().String(), adminTok(t))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestQualityCheckHandler_403_NonAdmin(t *testing.T) {
	contrib := &communitydomain.Contribution{ID: uuid.New(), Type: communitydomain.ContributionTypeNewWord}
	llm := &fakeLLM{}
	app := buildAdminApp(t, newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(contrib), llm, &fixedCounter{})

	resp := doAdmin(t, app, http.MethodPost, "/api/v2/admin/ai/check/"+contrib.ID.String(), userTok(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ensure nearLimitCounter is available in this file's test scope
var _ = time.Now
