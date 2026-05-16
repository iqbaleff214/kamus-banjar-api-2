package aihttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	aihttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/http"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	auth.Init("test-secret-that-is-long-enough-32ch")
}

// ─── fake LLM ─────────────────────────────────────────────────────────────────

type fakeLLM struct {
	response *domain.CompletionResponse
	err      error
}

func (f *fakeLLM) Complete(_ context.Context, _ domain.CompletionRequest) (*domain.CompletionResponse, error) {
	return f.response, f.err
}

// ─── fake counter ─────────────────────────────────────────────────────────────

type fixedCounter struct{ count int64 }

func (f *fixedCounter) Increment(_ context.Context, _ string, _ time.Duration) (int64, error) {
	f.count++
	return f.count, nil
}

// nearLimitCounter starts at n so the very next call exceeds limit.
func nearLimitCounter(n int64) *fixedCounter { return &fixedCounter{count: n} }

// ─── helpers ─────────────────────────────────────────────────────────────────

func newApp(llm domain.LLMClient, counter ratelimit.Counter) *fiber.App {
	svc := commands.NewTranslateService(llm, "test-model")
	h := aihttp.NewHandler(svc)
	app := fiber.New()
	aihttp.RegisterRoutes(app, h, (*aihttp.AdminHandler)(nil), counter)
	return app
}

func do(app *fiber.App, body any, token string) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/ai/translate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

func userToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "user")
	require.NoError(t, err)
	return tok
}

func decode(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&m))
	return m
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestTranslateHandler_401_NoToken(t *testing.T) {
	app := newApp(&fakeLLM{}, &fixedCounter{})
	resp := do(app, map[string]any{"text": "abah"}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestTranslateHandler_422_TooLong(t *testing.T) {
	llm := &fakeLLM{}
	app := newApp(llm, &fixedCounter{})
	resp := do(app, map[string]any{"text": strings.Repeat("a", 1001)}, userToken(t))
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestTranslateHandler_200(t *testing.T) {
	llm := &fakeLLM{
		response: &domain.CompletionResponse{
			Model:   "test-model",
			Content: `{"translation":"ayah","confidence":"high","notes":""}`,
		},
	}
	app := newApp(llm, &fixedCounter{})
	resp := do(app, map[string]any{"text": "abah"}, userToken(t))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decode(t, resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]any)
	assert.Equal(t, "abah", data["original"])
	assert.Equal(t, "ayah", data["translation"])
	assert.Equal(t, "hulu", data["dialect"])
	assert.Equal(t, "high", data["confidence"])
}

func TestTranslateHandler_503_LLMDown(t *testing.T) {
	llm := &fakeLLM{err: domain.ErrAIUnavailable}
	app := newApp(llm, &fixedCounter{})
	resp := do(app, map[string]any{"text": "abah"}, userToken(t))
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body := decode(t, resp)
	assert.False(t, body["success"].(bool))
}

func TestTranslateHandler_429_RateLimit(t *testing.T) {
	llm := &fakeLLM{
		response: &domain.CompletionResponse{
			Model:   "test-model",
			Content: `{"translation":"ayah","confidence":"high","notes":""}`,
		},
	}
	// start at 29 so 30th call passes (== limit), 31st exceeds it
	app := newApp(llm, nearLimitCounter(29))
	tok := userToken(t)

	first := do(app, map[string]any{"text": "abah"}, tok)
	require.Equal(t, http.StatusOK, first.StatusCode)

	second := do(app, map[string]any{"text": "abah"}, tok)
	assert.Equal(t, http.StatusTooManyRequests, second.StatusCode)
}
