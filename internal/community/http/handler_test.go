package communityhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/application/commands"
	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	communityhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/http"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

// ─── fakes ────────────────────────────────────────────────────────────────────

type fakeContribRepo struct {
	store map[uuid.UUID]*domain.Contribution
}

func newFakeContribRepo() *fakeContribRepo {
	return &fakeContribRepo{store: make(map[uuid.UUID]*domain.Contribution)}
}
func (r *fakeContribRepo) Create(_ context.Context, c *domain.Contribution) error {
	r.store[c.ID] = c
	return nil
}
func (r *fakeContribRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Contribution, error) {
	c, ok := r.store[id]
	if !ok {
		return nil, domain.ErrContributionNotFound
	}
	return c, nil
}
func (r *fakeContribRepo) FindByContributor(_ context.Context, userID uuid.UUID, _ domain.ContributionFilter, _, _ int) ([]*domain.Contribution, int, error) {
	var out []*domain.Contribution
	for _, c := range r.store {
		if c.ContributorID == userID {
			out = append(out, c)
		}
	}
	return out, len(out), nil
}
func (r *fakeContribRepo) FindAll(_ context.Context, _ domain.ContributionFilter, _, _ int) ([]*domain.Contribution, int, error) {
	var out []*domain.Contribution
	for _, c := range r.store {
		out = append(out, c)
	}
	return out, len(out), nil
}
func (r *fakeContribRepo) Update(_ context.Context, c *domain.Contribution) error {
	r.store[c.ID] = c
	return nil
}

type fakeWordChecker struct{ exists bool }

func (c *fakeWordChecker) Exists(_ context.Context, _ uuid.UUID) (bool, error) {
	return c.exists, nil
}

type fakeUserVerifier struct{ verified bool }

func (v *fakeUserVerifier) IsEmailVerified(_ context.Context, _ uuid.UUID) (bool, error) {
	return v.verified, nil
}

type fakeVoteRepo struct {
	store map[string]*domain.Vote
}

func newFakeVoteRepo() *fakeVoteRepo {
	return &fakeVoteRepo{store: make(map[string]*domain.Vote)}
}
func voteKey(u uuid.UUID, tt domain.VoteTargetType, t uuid.UUID) string {
	return u.String() + string(tt) + t.String()
}
func (r *fakeVoteRepo) Upsert(_ context.Context, v *domain.Vote) (*domain.Vote, error) {
	r.store[voteKey(v.UserID, v.TargetType, v.TargetID)] = v
	return v, nil
}
func (r *fakeVoteRepo) Delete(_ context.Context, u uuid.UUID, tt domain.VoteTargetType, t uuid.UUID) error {
	delete(r.store, voteKey(u, tt, t))
	return nil
}
func (r *fakeVoteRepo) FindByUserAndTarget(_ context.Context, u uuid.UUID, tt domain.VoteTargetType, t uuid.UUID) (*domain.Vote, error) {
	v, ok := r.store[voteKey(u, tt, t)]
	if !ok {
		return nil, domain.ErrVoteNotFound
	}
	return v, nil
}

type fakeBookmarkRepo struct {
	store map[string]*domain.Bookmark
}

func newFakeBookmarkRepo() *fakeBookmarkRepo {
	return &fakeBookmarkRepo{store: make(map[string]*domain.Bookmark)}
}
func bmKey(u, w uuid.UUID) string { return u.String() + w.String() }
func (r *fakeBookmarkRepo) Create(_ context.Context, b *domain.Bookmark) (*domain.Bookmark, error) {
	k := bmKey(b.UserID, b.WordID)
	if _, ok := r.store[k]; ok {
		return nil, domain.ErrBookmarkConflict
	}
	r.store[k] = b
	return b, nil
}
func (r *fakeBookmarkRepo) Delete(_ context.Context, u, w uuid.UUID) error {
	delete(r.store, bmKey(u, w))
	return nil
}
func (r *fakeBookmarkRepo) FindByUser(_ context.Context, u uuid.UUID, _, _ int) ([]*domain.Bookmark, int, error) {
	var out []*domain.Bookmark
	for _, b := range r.store {
		if b.UserID == u {
			out = append(out, b)
		}
	}
	return out, len(out), nil
}
func (r *fakeBookmarkRepo) Exists(_ context.Context, u, w uuid.UUID) (bool, error) {
	_, ok := r.store[bmKey(u, w)]
	return ok, nil
}

type fakeCommentRepo struct {
	store map[uuid.UUID]*domain.Comment
}

func newFakeCommentRepo() *fakeCommentRepo {
	return &fakeCommentRepo{store: make(map[uuid.UUID]*domain.Comment)}
}
func (r *fakeCommentRepo) Create(_ context.Context, c *domain.Comment) (*domain.Comment, error) {
	r.store[c.ID] = c
	return c, nil
}
func (r *fakeCommentRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	c, ok := r.store[id]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	return c, nil
}
func (r *fakeCommentRepo) FindByTarget(_ context.Context, _ domain.CommentTargetType, _ uuid.UUID, _, _ int) ([]*domain.Comment, int, error) {
	return nil, 0, nil
}
func (r *fakeCommentRepo) Update(_ context.Context, c *domain.Comment) (*domain.Comment, error) {
	r.store[c.ID] = c
	return c, nil
}
func (r *fakeCommentRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.store, id)
	return nil
}

// ─── counter helpers ──────────────────────────────────────────────────────────

type memCounter struct{ n int64 }

func (m *memCounter) Increment(_ context.Context, _ string, _ time.Duration) (int64, error) {
	m.n++
	return m.n, nil
}

func nearLimitCounter(n int64) *memCounter { return &memCounter{n: n} }

// ─── app builder ──────────────────────────────────────────────────────────────

func buildApp(t *testing.T, counter ratelimit.Counter, wordExists bool) *fiber.App {
	t.Helper()
	auth.Init("test-secret")

	contribSvc := commands.NewContributionService(
		newFakeContribRepo(),
		&fakeWordChecker{exists: wordExists},
		&fakeUserVerifier{verified: true},
	)
	voteSvc := commands.NewVoteService(
		newFakeVoteRepo(),
		&fakeWordChecker{exists: wordExists},
		&fakeWordChecker{exists: wordExists},
	)
	bookmarkSvc := commands.NewBookmarkService(newFakeBookmarkRepo(), &fakeWordChecker{exists: wordExists})
	commentSvc := commands.NewCommentService(newFakeCommentRepo(), &fakeWordChecker{exists: wordExists})

	app := fiber.New()
	h := communityhttp.NewHandler(contribSvc, voteSvc, bookmarkSvc, commentSvc)
	communityhttp.RegisterRoutes(app, h, counter)
	return app
}

func userToken(t *testing.T, role string) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(uuid.New().String(), role)
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

func decodeBody(t *testing.T, body io.ReadCloser, out any) {
	t.Helper()
	defer func() { _ = body.Close() }()
	require.NoError(t, json.NewDecoder(body).Decode(out))
}

// ─── contribution tests ───────────────────────────────────────────────────────

func TestSubmitContributionHandler_201(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New().String()
	resp := doJSON(t, app, http.MethodPost, "/api/v2/contributions", map[string]any{
		"type":           "new_definition",
		"target_word_id": wordID,
		"payload":        map[string]any{"meaning": "test"},
	}, userToken(t, "user"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestSubmitContributionHandler_429_RateLimit(t *testing.T) {
	app := buildApp(t, nearLimitCounter(10), true)
	resp := doJSON(t, app, http.MethodPost, "/api/v2/contributions", map[string]any{
		"type":    "new_word",
		"payload": map[string]any{"banjar": "test"},
	}, userToken(t, "user"))
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestApproveContributionHandler_403_NonAdmin(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	fakeID := uuid.New()
	resp := doJSON(t, app, http.MethodPatch, fmt.Sprintf("/api/v2/contributions/%s/approve", fakeID), nil, userToken(t, "user"))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRejectContributionHandler_422_MissingNote(t *testing.T) {
	// First submit, then try to reject without note
	app := buildApp(t, &memCounter{}, true)
	// Submit a contribution so we have an ID
	submitResp := doJSON(t, app, http.MethodPost, "/api/v2/contributions", map[string]any{
		"type":    "new_word",
		"payload": map[string]any{},
	}, userToken(t, "user"))
	require.Equal(t, http.StatusCreated, submitResp.StatusCode)

	var submitBody map[string]any
	decodeBody(t, submitResp.Body, &submitBody)
	id := submitBody["data"].(map[string]any)["id"].(string)

	// Admin rejects without note → 422
	resp := doJSON(t, app, http.MethodPatch, "/api/v2/contributions/"+id+"/reject",
		map[string]any{"note": ""}, userToken(t, "admin"))
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestWithdrawContributionHandler_409_NotPending(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	tok := userToken(t, "user")
	// Submit
	submitResp := doJSON(t, app, http.MethodPost, "/api/v2/contributions", map[string]any{
		"type":    "new_word",
		"payload": map[string]any{},
	}, tok)
	require.Equal(t, http.StatusCreated, submitResp.StatusCode)
	var submitBody map[string]any
	decodeBody(t, submitResp.Body, &submitBody)
	id := submitBody["data"].(map[string]any)["id"].(string)

	// Withdraw once → ok
	resp1 := doJSON(t, app, http.MethodPatch, "/api/v2/contributions/"+id+"/withdraw", nil, tok)
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	// Withdraw again → 409 (not pending)
	resp2 := doJSON(t, app, http.MethodPatch, "/api/v2/contributions/"+id+"/withdraw", nil, tok)
	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
}

// ─── vote tests ───────────────────────────────────────────────────────────────

func TestCastVoteHandler_201(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New()
	resp := doJSON(t, app, http.MethodPost, fmt.Sprintf("/api/v2/words/%s/votes", wordID), map[string]any{"value": "up"}, userToken(t, "user"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestRemoveVoteHandler_404_NoneExists(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New()
	resp := doJSON(t, app, http.MethodDelete, fmt.Sprintf("/api/v2/words/%s/votes", wordID), nil, userToken(t, "user"))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─── bookmark tests ───────────────────────────────────────────────────────────

func TestAddBookmarkHandler_409_Duplicate(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	tok := userToken(t, "user")
	wordID := uuid.New().String()

	doJSON(t, app, http.MethodPost, "/api/v2/bookmarks", map[string]any{"word_id": wordID}, tok)
	resp := doJSON(t, app, http.MethodPost, "/api/v2/bookmarks", map[string]any{"word_id": wordID}, tok)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// ─── comment tests ────────────────────────────────────────────────────────────

func TestPostCommentHandler_201(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New()
	resp := doJSON(t, app, http.MethodPost, fmt.Sprintf("/api/v2/words/%s/comments", wordID), map[string]any{"body": "nice word"}, userToken(t, "user"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestEditCommentHandler_403_NonOwner(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New()
	tok1 := userToken(t, "user")

	// Post comment as user1
	postResp := doJSON(t, app, http.MethodPost, fmt.Sprintf("/api/v2/words/%s/comments", wordID), map[string]any{"body": "hello"}, tok1)
	require.Equal(t, http.StatusCreated, postResp.StatusCode)
	var postBody map[string]any
	decodeBody(t, postResp.Body, &postBody)
	commentID := postBody["data"].(map[string]any)["id"].(string)

	// Edit as user2 → 403
	resp := doJSON(t, app, http.MethodPatch, "/api/v2/comments/"+commentID, map[string]any{"body": "changed"}, userToken(t, "user"))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteCommentHandler_204_Admin(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New()

	// Post comment as regular user
	postResp := doJSON(t, app, http.MethodPost, fmt.Sprintf("/api/v2/words/%s/comments", wordID), map[string]any{"body": "hello"}, userToken(t, "user"))
	require.Equal(t, http.StatusCreated, postResp.StatusCode)
	var postBody map[string]any
	decodeBody(t, postResp.Body, &postBody)
	commentID := postBody["data"].(map[string]any)["id"].(string)

	// Delete as admin → 204
	resp := doJSON(t, app, http.MethodDelete, "/api/v2/comments/"+commentID, nil, userToken(t, "admin"))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestFlagCommentHandler_200_Idempotent(t *testing.T) {
	app := buildApp(t, &memCounter{}, true)
	wordID := uuid.New()
	tok := userToken(t, "user")

	postResp := doJSON(t, app, http.MethodPost, fmt.Sprintf("/api/v2/words/%s/comments", wordID), map[string]any{"body": "hello"}, tok)
	require.Equal(t, http.StatusCreated, postResp.StatusCode)
	var postBody map[string]any
	decodeBody(t, postResp.Body, &postBody)
	commentID := postBody["data"].(map[string]any)["id"].(string)

	resp1 := doJSON(t, app, http.MethodPost, "/api/v2/comments/"+commentID+"/flag", nil, tok)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	resp2 := doJSON(t, app, http.MethodPost, "/api/v2/comments/"+commentID+"/flag", nil, tok)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}
