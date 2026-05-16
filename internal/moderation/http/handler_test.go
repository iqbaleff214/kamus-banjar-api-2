package moderationhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commands "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/application/commands"
	moderationhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/http"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
)

// ─── fakes ────────────────────────────────────────────────────────────────────

type fakeModerationRepo struct {
	contribs []*communitydomain.Contribution
	comments []*communitydomain.Comment
	users    []*identitydomain.User
	stats    moderationdomain.ModerationStats
}

func (f *fakeModerationRepo) GetPendingContributions(_ context.Context, _ *communitydomain.ContributionType, _, _ int) ([]*communitydomain.Contribution, int, error) {
	return f.contribs, len(f.contribs), nil
}
func (f *fakeModerationRepo) GetFlaggedComments(_ context.Context, _, _ int) ([]*communitydomain.Comment, int, error) {
	return f.comments, len(f.comments), nil
}
func (f *fakeModerationRepo) GetStats(_ context.Context) (*moderationdomain.ModerationStats, error) {
	return &f.stats, nil
}
func (f *fakeModerationRepo) CreateAuditLog(_ context.Context, _ *moderationdomain.AuditLog) error {
	return nil
}
func (f *fakeModerationRepo) ListUsers(_ context.Context, role, _ string, _ *bool, _, _ int) ([]*identitydomain.User, int, error) {
	if role == "" {
		return f.users, len(f.users), nil
	}
	var out []*identitydomain.User
	for _, u := range f.users {
		if string(u.Role) == role {
			out = append(out, u)
		}
	}
	return out, len(out), nil
}

type fakeContribWriter struct {
	store map[uuid.UUID]*communitydomain.Contribution
}

func newFakeContribWriter(cs ...*communitydomain.Contribution) *fakeContribWriter {
	m := make(map[uuid.UUID]*communitydomain.Contribution)
	for _, c := range cs {
		m[c.ID] = c
	}
	return &fakeContribWriter{store: m}
}
func (f *fakeContribWriter) FindByID(_ context.Context, id uuid.UUID) (*communitydomain.Contribution, error) {
	if c, ok := f.store[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeContribWriter) Update(_ context.Context, c *communitydomain.Contribution) error {
	f.store[c.ID] = c
	return nil
}

type fakeWordWriter struct {
	words map[uuid.UUID]*dictdomain.Word
}

func newFakeWordWriter() *fakeWordWriter {
	return &fakeWordWriter{words: make(map[uuid.UUID]*dictdomain.Word)}
}
func (f *fakeWordWriter) FindByID(_ context.Context, id uuid.UUID) (*dictdomain.Word, error) {
	if w, ok := f.words[id]; ok {
		return w, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeWordWriter) Create(_ context.Context, w *dictdomain.Word) error {
	f.words[w.ID] = w
	return nil
}
func (f *fakeWordWriter) Update(_ context.Context, w *dictdomain.Word) error {
	f.words[w.ID] = w
	return nil
}

type fakeUserWriter struct {
	store map[uuid.UUID]*identitydomain.User
}

func newFakeUserWriter(us ...*identitydomain.User) *fakeUserWriter {
	m := make(map[uuid.UUID]*identitydomain.User)
	for _, u := range us {
		m[u.ID] = u
	}
	return &fakeUserWriter{store: m}
}
func (f *fakeUserWriter) FindByID(_ context.Context, id uuid.UUID) (*identitydomain.User, error) {
	if u, ok := f.store[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeUserWriter) Update(_ context.Context, u *identitydomain.User) error {
	f.store[u.ID] = u
	return nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func buildApp(
	t *testing.T,
	modRepo *fakeModerationRepo,
	cw *fakeContribWriter,
	ww *fakeWordWriter,
	uw *fakeUserWriter,
) *fiber.App {
	t.Helper()
	auth.Init("test-secret")
	svc := commands.NewModerationService(modRepo, cw, ww, uw)
	h := moderationhttp.NewHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	moderationhttp.RegisterRoutes(app, h)
	return app
}

func adminToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "admin")
	require.NoError(t, err)
	return "Bearer " + tok
}

func userToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), "user")
	require.NoError(t, err)
	return "Bearer " + tok
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body any, token string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func newUser(role identitydomain.Role) *identitydomain.User {
	return &identitydomain.User{
		ID:        uuid.New(),
		Name:      "Test",
		Email:     "t@example.com",
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestModerationQueueHandler_200(t *testing.T) {
	contrib, _ := communitydomain.NewContribution(uuid.New(), communitydomain.ContributionTypeNewWord, nil, map[string]any{"banjar": "test"})
	modRepo := &fakeModerationRepo{contribs: []*communitydomain.Contribution{contrib}}
	app := buildApp(t, modRepo, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/moderation/queue", nil, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestFlaggedCommentsHandler_200(t *testing.T) {
	comment, _ := communitydomain.NewComment(uuid.New(), communitydomain.CommentTargetWord, uuid.New(), "great word!")
	comment.Flag()
	modRepo := &fakeModerationRepo{comments: []*communitydomain.Comment{comment}}
	app := buildApp(t, modRepo, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/moderation/flags", nil, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestModerationStatsHandler_200(t *testing.T) {
	modRepo := &fakeModerationRepo{stats: moderationdomain.ModerationStats{PendingContributions: 3, FlaggedComments: 1}}
	app := buildApp(t, modRepo, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/moderation/stats", nil, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	defer func() { _ = resp.Body.Close() }()
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(3), data["pending_contributions"])
}

func TestAdminUsersHandler_FilterByRole(t *testing.T) {
	admin := newUser(identitydomain.RoleAdmin)
	user := newUser(identitydomain.RoleUser)
	modRepo := &fakeModerationRepo{users: []*identitydomain.User{admin, user}}
	app := buildApp(t, modRepo, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/users?role=admin", nil, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	defer func() { _ = resp.Body.Close() }()
	data := body["data"].([]any)
	assert.Equal(t, 1, len(data))
}

func TestBanUserHandler_200(t *testing.T) {
	target := newUser(identitydomain.RoleUser)
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	resp := doJSON(t, app, http.MethodPatch, "/api/v2/admin/users/"+target.ID.String()+"/ban",
		map[string]any{"reason": "spammer"}, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBanUserHandler_403_BanAdmin(t *testing.T) {
	adminTarget := newUser(identitydomain.RoleAdmin)
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(adminTarget))

	resp := doJSON(t, app, http.MethodPatch, "/api/v2/admin/users/"+adminTarget.ID.String()+"/ban",
		map[string]any{"reason": "test"}, adminToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestChangeRoleHandler_422_InvalidRole(t *testing.T) {
	target := newUser(identitydomain.RoleUser)
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	resp := doJSON(t, app, http.MethodPatch, "/api/v2/admin/users/"+target.ID.String()+"/role",
		map[string]any{"role": "superuser"}, adminToken(t))
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAdminEndpoints_403_NonAdmin(t *testing.T) {
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/moderation/stats", nil, userToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminEndpoints_401_NoToken(t *testing.T) {
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/moderation/stats", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ─── GetUser / UnbanUser tests ────────────────────────────────────────────────

func TestGetUserHandler_200(t *testing.T) {
	target := newUser(identitydomain.RoleUser)
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/users/"+target.ID.String(), nil, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, true, body["success"])
}

func TestGetUserHandler_404(t *testing.T) {
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/users/"+uuid.New().String(), nil, adminToken(t))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetUserHandler_404_InvalidID(t *testing.T) {
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter())

	resp := doJSON(t, app, http.MethodGet, "/api/v2/admin/users/not-a-uuid", nil, adminToken(t))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUnbanUserHandler_200(t *testing.T) {
	target := newUser(identitydomain.RoleUser)
	target.IsActive = false
	app := buildApp(t, &fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	resp := doJSON(t, app, http.MethodPatch, "/api/v2/admin/users/"+target.ID.String()+"/unban", nil, adminToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
