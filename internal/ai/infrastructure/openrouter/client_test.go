package openrouter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/infrastructure/openrouter"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validResponse() map[string]any {
	return map[string]any{
		"model": "mistralai/mistral-7b-instruct:free",
		"choices": []map[string]any{
			{
				"message": map[string]any{
					"role":    "assistant",
					"content": `{"translation":"dia tidak bisa pergi ke pasar","confidence":"high","notes":"kada=tidak"}`,
				},
			},
		},
	}
}

func newTestClient(t *testing.T, server *httptest.Server) *openrouter.Client {
	t.Helper()
	cfg := &config.Config{
		OpenRouterAPIKey:  "test-key",
		OpenRouterBaseURL: server.URL,
		OpenRouterModel:   "mistralai/mistral-7b-instruct:free",
	}
	c, err := openrouter.New(cfg)
	require.NoError(t, err)
	return c
}

func TestOpenRouterClient_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "Kamus Banjar API", r.Header.Get("X-Title"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(validResponse())
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	resp, err := client.Complete(context.Background(), domain.CompletionRequest{
		Model:    "mistralai/mistral-7b-instruct:free",
		Messages: []domain.Message{{Role: "user", Content: "translate: abah"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "mistralai/mistral-7b-instruct:free", resp.Model)
	assert.NotEmpty(t, resp.Content)
}

func TestOpenRouterClient_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Complete(context.Background(), domain.CompletionRequest{
		Messages: []domain.Message{{Role: "user", Content: "test"}},
	})
	assert.True(t, errors.Is(err, domain.ErrAIUnavailable))
}

func TestOpenRouterClient_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// close connection immediately without response
		if h, ok := w.(http.Hijacker); ok {
			conn, _, _ := h.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Complete(context.Background(), domain.CompletionRequest{
		Messages: []domain.Message{{Role: "user", Content: "test"}},
	})
	assert.True(t, errors.Is(err, domain.ErrAIUnavailable))
}

func TestOpenRouterClient_MissingAPIKey(t *testing.T) {
	cfg := &config.Config{OpenRouterAPIKey: "", OpenRouterBaseURL: "http://localhost", OpenRouterModel: "x"}
	_, err := openrouter.New(cfg)
	require.Error(t, err)
}
