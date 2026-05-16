package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	aidomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
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
	return nil, 0, nil
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
func (f *fakeWordAccessor) FindByBanjar(_ context.Context, banjar string) (*dictdomain.Word, error) {
	for _, w := range f.store {
		if w.Banjar == banjar {
			return w, nil
		}
	}
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

type stubLLM struct {
	resp *aidomain.CompletionResponse
	err  error
}

func (f *stubLLM) Complete(_ context.Context, _ aidomain.CompletionRequest) (*aidomain.CompletionResponse, error) {
	return f.resp, f.err
}

func newWord() *dictdomain.Word {
	w, _ := dictdomain.NewWord("urang", dictdomain.WordClassNomina, dictdomain.DialectHulu)
	return w
}

func newSvc(aiRepo *fakeAIRepo, wordRepo *fakeWordAccessor, contribRepo *fakeContribReader, llm aidomain.LLMClient) *commands.EnrichmentService {
	return commands.NewEnrichmentService(aiRepo, wordRepo, contribRepo, &fakeAuditLogWriter{}, llm, "test-model")
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestTriggerEnrichment_WordNotFound(t *testing.T) {
	svc := newSvc(newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(), &stubLLM{})

	_, err := svc.TriggerDefinitionEnrichment(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, commands.ErrWordNotFound)
}

func TestTriggerEnrichment_LLMFails_StatusFailed(t *testing.T) {
	word := newWord()
	aiRepo := newFakeAIRepo()
	llm := &stubLLM{err: aidomain.ErrAIUnavailable}
	svc := newSvc(aiRepo, newFakeWordAccessor(word), newFakeContribReader(), llm)

	req, err := svc.TriggerDefinitionEnrichment(context.Background(), uuid.New(), word.ID)
	require.NoError(t, err)
	assert.Equal(t, aidomain.AIRequestStatusPending, req.Status)

	// give goroutine time to run
	assert.Eventually(t, func() bool {
		r, _ := aiRepo.FindByID(context.Background(), req.ID)
		return r != nil && r.Status == aidomain.AIRequestStatusFailed
	}, 2*time.Second, 10*time.Millisecond)
}

func TestTriggerEnrichment_LLMSucceeds_StatusCompleted(t *testing.T) {
	word := newWord()
	aiRepo := newFakeAIRepo()
	llm := &stubLLM{resp: &aidomain.CompletionResponse{
		Model:   "test-model",
		Content: `{"definition":"orang, manusia"}`,
	}}
	svc := newSvc(aiRepo, newFakeWordAccessor(word), newFakeContribReader(), llm)

	req, err := svc.TriggerDefinitionEnrichment(context.Background(), uuid.New(), word.ID)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		r, _ := aiRepo.FindByID(context.Background(), req.ID)
		return r != nil && r.Status == aidomain.AIRequestStatusCompleted
	}, 2*time.Second, 10*time.Millisecond)
}

func TestTriggerQualityCheck_ContributionNotFound(t *testing.T) {
	svc := newSvc(newFakeAIRepo(), newFakeWordAccessor(), newFakeContribReader(), &stubLLM{})

	_, err := svc.TriggerQualityCheck(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, commands.ErrContributionNotFound)
}

func TestTriggerQualityCheck_StoresAdvisoryOutput(t *testing.T) {
	contrib, _ := communitydomain.NewContribution(uuid.New(), communitydomain.ContributionTypeNewWord, nil, map[string]any{"banjar": "urang"})
	aiRepo := newFakeAIRepo()
	llm := &stubLLM{resp: &aidomain.CompletionResponse{
		Model:   "test-model",
		Content: `{"score":90,"notes":"good","issues":[]}`,
	}}
	svc := newSvc(aiRepo, newFakeWordAccessor(), newFakeContribReader(contrib), llm)

	req, err := svc.TriggerQualityCheck(context.Background(), uuid.New(), contrib.ID)
	require.NoError(t, err)
	assert.Equal(t, aidomain.AIRequestTypeQualityCheck, req.Type)

	assert.Eventually(t, func() bool {
		r, _ := aiRepo.FindByID(context.Background(), req.ID)
		return r != nil && r.Status == aidomain.AIRequestStatusCompleted
	}, 2*time.Second, 10*time.Millisecond)
}

func TestApproveAIRequest_QualityCheck_Conflict(t *testing.T) {
	aiRepo := newFakeAIRepo()
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeQualityCheck, nil, nil, uuid.New(), "m", "p")
	_ = req.MarkCompleted(map[string]any{}, map[string]any{"score": 80})
	_ = aiRepo.Create(context.Background(), req)

	svc := newSvc(aiRepo, newFakeWordAccessor(), newFakeContribReader(), &stubLLM{})
	_, err := svc.ApproveAIRequest(context.Background(), uuid.New(), req.ID)
	assert.ErrorIs(t, err, commands.ErrAIRequestConflict)
}

func TestApproveAIRequest_AlreadyApproved_Conflict(t *testing.T) {
	wordID := uuid.New()
	aiRepo := newFakeAIRepo()
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, &wordID, nil, uuid.New(), "m", "p")
	_ = req.MarkCompleted(map[string]any{}, map[string]any{"definition": "meaning"})
	_ = req.Approve(uuid.New())
	_ = aiRepo.Create(context.Background(), req)

	word := newWord()
	word.ID = wordID
	svc := newSvc(aiRepo, newFakeWordAccessor(word), newFakeContribReader(), &stubLLM{})
	_, err := svc.ApproveAIRequest(context.Background(), uuid.New(), req.ID)
	assert.ErrorIs(t, err, commands.ErrAIRequestConflict)
}

func TestApproveAIRequest_MergesOutput(t *testing.T) {
	word := newWord()
	wordID := word.ID
	aiRepo := newFakeAIRepo()
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, &wordID, nil, uuid.New(), "m", "p")
	_ = req.MarkCompleted(map[string]any{}, map[string]any{"definition": "manusia, orang"})
	_ = aiRepo.Create(context.Background(), req)

	wordAccessor := newFakeWordAccessor(word)
	svc := newSvc(aiRepo, wordAccessor, newFakeContribReader(), &stubLLM{})

	approved, err := svc.ApproveAIRequest(context.Background(), uuid.New(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, aidomain.AIReviewStatusApproved, approved.ReviewStatus)

	updated, _ := wordAccessor.FindByID(context.Background(), word.ID)
	assert.Len(t, updated.Definitions, 1)
	assert.Equal(t, "manusia, orang", updated.Definitions[0].Meaning)
}

func TestRejectAIRequest_AlreadyRejected_Conflict(t *testing.T) {
	aiRepo := newFakeAIRepo()
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, nil, nil, uuid.New(), "m", "p")
	_ = req.MarkCompleted(map[string]any{}, map[string]any{})
	_ = req.Reject(uuid.New())
	_ = aiRepo.Create(context.Background(), req)

	svc := newSvc(aiRepo, newFakeWordAccessor(), newFakeContribReader(), &stubLLM{})
	_, err := svc.RejectAIRequest(context.Background(), uuid.New(), req.ID)
	assert.ErrorIs(t, err, commands.ErrAIRequestConflict)
}
