package identityhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	identityhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/http"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	auth.Init("test-secret-that-is-long-enough-32chars")
}

// --- in-memory fakes (same as commands test) ---

type fakeUserRepo struct {
	users   map[string]*domain.User
	byEmail map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User), byEmail: make(map[string]*domain.User)}
}
func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if _, exists := r.byEmail[u.Email]; exists {
		return domain.ErrEmailConflict
	}
	r.users[u.ID.String()] = u
	r.byEmail[u.Email] = u
	return nil
}
func (r *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}
func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}
func (r *fakeUserRepo) Update(_ context.Context, u *domain.User) error {
	r.users[u.ID.String()] = u
	r.byEmail[u.Email] = u
	return nil
}

type fakeTokenStore struct{ data map[string]string }

func newFakeTokenStore() *fakeTokenStore { return &fakeTokenStore{data: make(map[string]string)} }
func (s *fakeTokenStore) StoreRefreshToken(_ context.Context, t, uid string, _ time.Duration) error {
	s.data["refresh:"+t] = uid; return nil
}
func (s *fakeTokenStore) GetUserIDByRefreshToken(_ context.Context, t string) (string, error) {
	v, ok := s.data["refresh:"+t]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}
func (s *fakeTokenStore) RevokeRefreshToken(_ context.Context, t string) error {
	delete(s.data, "refresh:"+t); return nil
}
func (s *fakeTokenStore) StoreVerificationToken(_ context.Context, t, uid string, _ time.Duration) error {
	s.data["verify:"+t] = uid; return nil
}
func (s *fakeTokenStore) GetUserIDByVerificationToken(_ context.Context, t string) (string, error) {
	v, ok := s.data["verify:"+t]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}
func (s *fakeTokenStore) DeleteVerificationToken(_ context.Context, t string) error {
	delete(s.data, "verify:"+t); return nil
}
func (s *fakeTokenStore) StoreResetToken(_ context.Context, t, uid string, _ time.Duration) error {
	s.data["reset:"+t] = uid; return nil
}
func (s *fakeTokenStore) GetUserIDByResetToken(_ context.Context, t string) (string, error) {
	v, ok := s.data["reset:"+t]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}
func (s *fakeTokenStore) DeleteResetToken(_ context.Context, t string) error {
	delete(s.data, "reset:"+t); return nil
}

func newApp() (*fiber.App, *mailer.MockMailer) {
	repo := newFakeUserRepo()
	store := newFakeTokenStore()
	mock := &mailer.MockMailer{}
	svc := commands.NewService(repo, store, mock)
	h := identityhttp.NewHandler(svc)

	app := fiber.New()
	identityhttp.RegisterRoutes(app, h)
	return app, mock
}

func doRequest(app *fiber.App, method, path string, body any, token string) (int, map[string]any) {
	var r *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	} else {
		r = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		return 0, nil
	}
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return resp.StatusCode, m
}

func TestRegisterHandler_201(t *testing.T) {
	app, _ := newApp()
	status, body := doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	assert.Equal(t, 201, status)
	assert.True(t, body["success"].(bool))
}

func TestLoginHandler_200_TokensPresent(t *testing.T) {
	app, _ := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")

	status, body := doRequest(app, "POST", "/api/v2/auth/login", map[string]any{
		"email": "alice@example.com", "password": "password123",
	}, "")
	require.Equal(t, 200, status)
	data := body["data"].(map[string]any)
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.Equal(t, float64(900), data["expires_in"])
}

func TestLoginHandler_401_WrongPassword(t *testing.T) {
	app, _ := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")

	status, _ := doRequest(app, "POST", "/api/v2/auth/login", map[string]any{
		"email": "alice@example.com", "password": "wrongpassword",
	}, "")
	assert.Equal(t, 401, status)
}

func TestRefreshHandler_200(t *testing.T) {
	app, _ := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	_, loginBody := doRequest(app, "POST", "/api/v2/auth/login", map[string]any{
		"email": "alice@example.com", "password": "password123",
	}, "")
	refreshToken := loginBody["data"].(map[string]any)["refresh_token"].(string)

	status, body := doRequest(app, "POST", "/api/v2/auth/refresh", map[string]any{
		"refresh_token": refreshToken,
	}, "")
	require.Equal(t, 200, status)
	data := body["data"].(map[string]any)
	assert.NotEmpty(t, data["access_token"])
}

func TestLogoutHandler_204(t *testing.T) {
	app, _ := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	_, loginBody := doRequest(app, "POST", "/api/v2/auth/login", map[string]any{
		"email": "alice@example.com", "password": "password123",
	}, "")
	refreshToken := loginBody["data"].(map[string]any)["refresh_token"].(string)

	status, _ := doRequest(app, "POST", "/api/v2/auth/logout", map[string]any{
		"refresh_token": refreshToken,
	}, "")
	assert.Equal(t, 204, status)
}

func TestVerifyEmailHandler_204(t *testing.T) {
	app, mock := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	require.NotEmpty(t, mock.VerificationCalls)
	token := mock.VerificationCalls[0].Token

	status, _ := doRequest(app, "POST", "/api/v2/auth/verify-email", map[string]any{
		"token": token,
	}, "")
	assert.Equal(t, 204, status)
}

func TestForgotPasswordHandler_204(t *testing.T) {
	app, _ := newApp()
	status, _ := doRequest(app, "POST", "/api/v2/auth/forgot-password", map[string]any{
		"email": "anyone@example.com",
	}, "")
	assert.Equal(t, 204, status)
}

func TestResetPasswordHandler_204(t *testing.T) {
	app, mock := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	_, _ = doRequest(app, "POST", "/api/v2/auth/forgot-password", map[string]any{
		"email": "alice@example.com",
	}, "")
	require.NotEmpty(t, mock.ResetCalls)
	token := mock.ResetCalls[0].Token

	status, _ := doRequest(app, "POST", "/api/v2/auth/reset-password", map[string]any{
		"token": token, "password": "newpassword123", "password_confirmation": "newpassword123",
	}, "")
	assert.Equal(t, 204, status)
}

func TestMeHandler_RequiresAuth(t *testing.T) {
	app, _ := newApp()
	status, _ := doRequest(app, "GET", "/api/v2/me/", nil, "")
	assert.Equal(t, 401, status)
}

func TestUpdateProfileHandler_200(t *testing.T) {
	app, _ := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	_, loginBody := doRequest(app, "POST", "/api/v2/auth/login", map[string]any{
		"email": "alice@example.com", "password": "password123",
	}, "")
	accessToken := loginBody["data"].(map[string]any)["access_token"].(string)

	status, body := doRequest(app, "PUT", "/api/v2/me/", map[string]any{"name": "Alice Updated"}, accessToken)
	require.Equal(t, 200, status)
	data := body["data"].(map[string]any)
	assert.Equal(t, "Alice Updated", data["name"])
}

func TestChangePasswordHandler_204(t *testing.T) {
	app, _ := newApp()
	_, _ = doRequest(app, "POST", "/api/v2/auth/register", map[string]any{
		"name": "Alice", "email": "alice@example.com",
		"password": "password123", "password_confirmation": "password123",
	}, "")
	_, loginBody := doRequest(app, "POST", "/api/v2/auth/login", map[string]any{
		"email": "alice@example.com", "password": "password123",
	}, "")
	accessToken := loginBody["data"].(map[string]any)["access_token"].(string)

	status, _ := doRequest(app, "PUT", "/api/v2/me/password", map[string]any{
		"current_password": "password123", "new_password": "newpassword123",
	}, accessToken)
	assert.Equal(t, 204, status)
}
