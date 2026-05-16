package ratelimit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

// ─── in-memory counter ────────────────────────────────────────────────────────

type memCounter struct {
	mu     sync.Mutex
	counts map[string]int64
}

func newMemCounter() *memCounter {
	return &memCounter{counts: make(map[string]int64)}
}

func (m *memCounter) Increment(_ context.Context, key string, _ time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[key]++
	return m.counts[key], nil
}

func (m *memCounter) Reset(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.counts, key)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newApp(counter ratelimit.Counter, limit int) *fiber.App {
	app := fiber.New()
	app.Use(ratelimit.Limiter(counter, func(c *fiber.Ctx) string { return "test-key" }, limit, time.Hour))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	return app
}

func get(app *fiber.App) int {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	return resp.StatusCode
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestRateLimiter_UnderLimit(t *testing.T) {
	counter := newMemCounter()
	app := newApp(counter, 3)
	for range 3 {
		assert.Equal(t, http.StatusOK, get(app))
	}
}

func TestRateLimiter_AtLimit(t *testing.T) {
	counter := newMemCounter()
	app := newApp(counter, 3)
	for range 3 {
		require.Equal(t, http.StatusOK, get(app))
	}
	assert.Equal(t, http.StatusTooManyRequests, get(app))
}

func TestRateLimiter_WindowReset(t *testing.T) {
	counter := newMemCounter()
	app := newApp(counter, 2)
	require.Equal(t, http.StatusOK, get(app))
	require.Equal(t, http.StatusOK, get(app))
	require.Equal(t, http.StatusTooManyRequests, get(app))

	// Simulate window reset by clearing the counter
	counter.Reset("test-key")
	assert.Equal(t, http.StatusOK, get(app))
}

func TestRateLimiter_FailOpen(t *testing.T) {
	// errorCounter always returns an error — middleware should pass through
	app := fiber.New()
	app.Use(ratelimit.Limiter(&errorCounter{}, func(_ *fiber.Ctx) string { return "k" }, 1, time.Hour))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	assert.Equal(t, http.StatusOK, get(app))
}

type errorCounter struct{}

func (e *errorCounter) Increment(_ context.Context, _ string, _ time.Duration) (int64, error) {
	return 0, fiber.ErrInternalServerError
}
