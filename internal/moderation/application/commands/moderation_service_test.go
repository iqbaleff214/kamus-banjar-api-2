package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	commands "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/application/commands"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

// ─── fakes ───────────────────────────────────────────────────────────────────

type fakeModerationRepo struct {
	auditLogs []*moderationdomain.AuditLog
}

func (f *fakeModerationRepo) GetPendingContributions(_ context.Context, _ *communitydomain.ContributionType, _, _ int) ([]*communitydomain.Contribution, int, error) {
	return nil, 0, nil
}
func (f *fakeModerationRepo) GetFlaggedComments(_ context.Context, _, _ int) ([]*communitydomain.Comment, int, error) {
	return nil, 0, nil
}
func (f *fakeModerationRepo) GetStats(_ context.Context) (*moderationdomain.ModerationStats, error) {
	return &moderationdomain.ModerationStats{}, nil
}
func (f *fakeModerationRepo) CreateAuditLog(_ context.Context, log *moderationdomain.AuditLog) error {
	f.auditLogs = append(f.auditLogs, log)
	return nil
}
func (f *fakeModerationRepo) ListUsers(_ context.Context, _, _ string, _ *bool, _, _ int) ([]*identitydomain.User, int, error) {
	return nil, 0, nil
}

type fakeContribWriter struct {
	contributions map[uuid.UUID]*communitydomain.Contribution
}

func newFakeContribWriter(contribs ...*communitydomain.Contribution) *fakeContribWriter {
	m := make(map[uuid.UUID]*communitydomain.Contribution)
	for _, c := range contribs {
		m[c.ID] = c
	}
	return &fakeContribWriter{contributions: m}
}

func (f *fakeContribWriter) FindByID(_ context.Context, id uuid.UUID) (*communitydomain.Contribution, error) {
	if c, ok := f.contributions[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeContribWriter) Update(_ context.Context, c *communitydomain.Contribution) error {
	f.contributions[c.ID] = c
	return nil
}

type fakeWordWriter struct {
	words   map[uuid.UUID]*dictdomain.Word
	created []*dictdomain.Word
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
	f.created = append(f.created, w)
	return nil
}
func (f *fakeWordWriter) Update(_ context.Context, w *dictdomain.Word) error {
	f.words[w.ID] = w
	return nil
}

type fakeUserWriter struct {
	users map[uuid.UUID]*identitydomain.User
}

func newFakeUserWriter(users ...*identitydomain.User) *fakeUserWriter {
	m := make(map[uuid.UUID]*identitydomain.User)
	for _, u := range users {
		m[u.ID] = u
	}
	return &fakeUserWriter{users: m}
}

func (f *fakeUserWriter) FindByID(_ context.Context, id uuid.UUID) (*identitydomain.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeUserWriter) Update(_ context.Context, u *identitydomain.User) error {
	f.users[u.ID] = u
	return nil
}

// ─── helper ───────────────────────────────────────────────────────────────────

func newPendingContrib(ctype communitydomain.ContributionType, payload map[string]any) *communitydomain.Contribution {
	c, _ := communitydomain.NewContribution(uuid.New(), ctype, nil, payload)
	return c
}

func newPendingContribWithTarget(ctype communitydomain.ContributionType, targetWordID uuid.UUID, payload map[string]any) *communitydomain.Contribution {
	c, _ := communitydomain.NewContribution(uuid.New(), ctype, &targetWordID, payload)
	return c
}

func newTestUser(role identitydomain.Role, isActive bool) *identitydomain.User {
	id := uuid.New()
	return &identitydomain.User{
		ID:        id,
		Name:      "Test User",
		Email:     "test@example.com",
		Role:      role,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func newSvc(modRepo *fakeModerationRepo, cw *fakeContribWriter, ww *fakeWordWriter, uw *fakeUserWriter) *commands.ModerationService {
	return commands.NewModerationService(modRepo, cw, ww, uw)
}

// ─── ApproveContribution tests ────────────────────────────────────────────────

func TestApproveContribution_NewWord_CreatesWord(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	words := newFakeWordWriter()
	payload := map[string]any{
		"banjar":     "kambang",
		"word_class": "n",
		"definitions": []any{
			map[string]any{"meaning": "bunga"},
		},
	}
	contrib := newPendingContrib(communitydomain.ContributionTypeNewWord, payload)
	contribs := newFakeContribWriter(contrib)
	svc := newSvc(modRepo, contribs, words, newFakeUserWriter())

	if err := svc.ApproveContribution(context.Background(), uuid.New(), contrib.ID, "good word"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words.created) != 1 {
		t.Errorf("expected 1 word created, got %d", len(words.created))
	}
	if words.created[0].Banjar != "kambang" {
		t.Errorf("expected banjar=kambang, got %q", words.created[0].Banjar)
	}
	if len(modRepo.auditLogs) != 1 {
		t.Errorf("expected 1 audit log, got %d", len(modRepo.auditLogs))
	}
}

func TestApproveContribution_NewDefinition_AddsDefinition(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	existingWord, _ := dictdomain.NewWord("kambang", dictdomain.WordClassNomina, dictdomain.DialectHulu)
	words := newFakeWordWriter()
	words.words[existingWord.ID] = existingWord

	payload := map[string]any{"meaning": "mekar"}
	contrib := newPendingContribWithTarget(communitydomain.ContributionTypeNewDefinition, existingWord.ID, payload)
	contribs := newFakeContribWriter(contrib)
	svc := newSvc(modRepo, contribs, words, newFakeUserWriter())

	if err := svc.ApproveContribution(context.Background(), uuid.New(), contrib.ID, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated := words.words[existingWord.ID]
	if len(updated.Definitions) != 1 {
		t.Errorf("expected 1 definition, got %d", len(updated.Definitions))
	}
}

func TestApproveContribution_NonPending_Conflict(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	contrib := newPendingContrib(communitydomain.ContributionTypeNewWord, map[string]any{
		"banjar": "test", "word_class": "n", "definitions": []any{},
	})
	adminID := uuid.New()
	_ = contrib.Approve(adminID, "pre-approved")

	contribs := newFakeContribWriter(contrib)
	svc := newSvc(modRepo, contribs, newFakeWordWriter(), newFakeUserWriter())

	err := svc.ApproveContribution(context.Background(), adminID, contrib.ID, "again")
	if !errors.Is(err, commands.ErrContributionConflict) {
		t.Errorf("expected ErrContributionConflict, got %v", err)
	}
}

func TestApproveContribution_WritesAuditLog(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	payload := map[string]any{
		"banjar":      "urang",
		"word_class":  "n",
		"definitions": []any{map[string]any{"meaning": "orang"}},
	}
	contrib := newPendingContrib(communitydomain.ContributionTypeNewWord, payload)
	svc := newSvc(modRepo, newFakeContribWriter(contrib), newFakeWordWriter(), newFakeUserWriter())

	adminID := uuid.New()
	_ = svc.ApproveContribution(context.Background(), adminID, contrib.ID, "ok")

	if len(modRepo.auditLogs) == 0 {
		t.Fatal("expected audit log to be written")
	}
	if modRepo.auditLogs[0].Action != moderationdomain.ActionApproveContribution {
		t.Errorf("unexpected action: %v", modRepo.auditLogs[0].Action)
	}
}

// ─── RejectContribution tests ─────────────────────────────────────────────────

func TestRejectContribution_EmptyNote(t *testing.T) {
	contrib := newPendingContrib(communitydomain.ContributionTypeNewWord, map[string]any{})
	svc := newSvc(&fakeModerationRepo{}, newFakeContribWriter(contrib), newFakeWordWriter(), newFakeUserWriter())

	err := svc.RejectContribution(context.Background(), uuid.New(), contrib.ID, "")
	if !errors.Is(err, commands.ErrNoteRequired) {
		t.Errorf("expected ErrNoteRequired, got %v", err)
	}
}

func TestRejectContribution_WritesAuditLog(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	contrib := newPendingContrib(communitydomain.ContributionTypeNewWord, map[string]any{})
	svc := newSvc(modRepo, newFakeContribWriter(contrib), newFakeWordWriter(), newFakeUserWriter())

	adminID := uuid.New()
	_ = svc.RejectContribution(context.Background(), adminID, contrib.ID, "bad content")

	if len(modRepo.auditLogs) == 0 {
		t.Fatal("expected audit log")
	}
	if modRepo.auditLogs[0].Action != moderationdomain.ActionRejectContribution {
		t.Errorf("unexpected action: %v", modRepo.auditLogs[0].Action)
	}
}

// ─── BanUser tests ────────────────────────────────────────────────────────────

func TestBanUser_CannotBanAdmin(t *testing.T) {
	adminTarget := newTestUser(identitydomain.RoleAdmin, true)
	svc := newSvc(&fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(adminTarget))

	_, err := svc.BanUser(context.Background(), uuid.New(), adminTarget.ID, "reason")
	if !errors.Is(err, moderationdomain.ErrCannotBanAdmin) {
		t.Errorf("expected ErrCannotBanAdmin, got %v", err)
	}
}

func TestBanUser_WritesAuditLog(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	target := newTestUser(identitydomain.RoleUser, true)
	svc := newSvc(modRepo, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	adminID := uuid.New()
	if _, err := svc.BanUser(context.Background(), adminID, target.ID, "spammer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modRepo.auditLogs) == 0 {
		t.Fatal("expected audit log")
	}
}

func TestUnbanUser_NotBanned_Idempotent(t *testing.T) {
	target := newTestUser(identitydomain.RoleUser, true)
	svc := newSvc(&fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	u, err := svc.UnbanUser(context.Background(), uuid.New(), target.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !u.IsActive {
		t.Errorf("expected user to remain active after unban of already-active user")
	}
}

// ─── ChangeUserRole tests ─────────────────────────────────────────────────────

func TestChangeUserRole_SelfDemotion_Forbidden(t *testing.T) {
	adminID := uuid.New()
	admin := &identitydomain.User{ID: adminID, Role: identitydomain.RoleAdmin, IsActive: true}
	svc := newSvc(&fakeModerationRepo{}, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(admin))

	_, err := svc.ChangeUserRole(context.Background(), adminID, adminID, identitydomain.RoleUser)
	if !errors.Is(err, moderationdomain.ErrCannotSelfDemote) {
		t.Errorf("expected ErrCannotSelfDemote, got %v", err)
	}
}

func TestChangeUserRole_WritesAuditLog(t *testing.T) {
	modRepo := &fakeModerationRepo{}
	target := newTestUser(identitydomain.RoleUser, true)
	svc := newSvc(modRepo, newFakeContribWriter(), newFakeWordWriter(), newFakeUserWriter(target))

	adminID := uuid.New()
	if _, err := svc.ChangeUserRole(context.Background(), adminID, target.ID, identitydomain.RoleAdmin); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modRepo.auditLogs) == 0 {
		t.Fatal("expected audit log")
	}
	if modRepo.auditLogs[0].Action != moderationdomain.ActionChangeRole {
		t.Errorf("unexpected action: %v", modRepo.auditLogs[0].Action)
	}
}
