package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/middleware"
)

func newApp() *fiber.App {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	return app
}

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	app := newApp()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
}

func TestRequestIDMiddleware_PreservesClientID(t *testing.T) {
	app := newApp()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "my-client-id-123")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, "my-client-id-123", resp.Header.Get("X-Request-ID"))
}
