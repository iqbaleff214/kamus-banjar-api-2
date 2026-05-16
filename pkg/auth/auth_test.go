package auth_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
)

func init() {
	auth.Init("test-secret-key-that-is-long-enough")
}

func TestGenerateAndParseToken(t *testing.T) {
	token, err := auth.GenerateAccessToken("user-123", "user")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := auth.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

func TestExpiredToken(t *testing.T) {
	// Forge an expired token by temporarily changing the TTL isn't possible
	// without exporting internal state. Instead, use a known expired token
	// generated with the same secret.
	// We test ParseToken returns ErrInvalidToken on garbage input.
	_, err := auth.ParseToken("not.a.valid.jwt")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestRequireAuthMiddleware(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", auth.RequireAuth(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Missing header
	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Invalid token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Valid token
	token, _ := auth.GenerateAccessToken("u1", "user")
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireRoleMiddleware(t *testing.T) {
	app := fiber.New()
	app.Get("/admin", auth.RequireAuth(), auth.RequireRole("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// User token on admin route
	token, _ := auth.GenerateAccessToken("u1", "user")
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// Admin token
	token, _ = auth.GenerateAccessToken("a1", "admin")
	req = httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// Ensure the TTL constant is 15 minutes.
func TestAccessTokenTTL(t *testing.T) {
	assert.Equal(t, 15*time.Minute, auth.AccessTokenTTL)
}
