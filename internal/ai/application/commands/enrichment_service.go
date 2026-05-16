package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	aidomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

type EnrichmentService struct {
	aiRepo      aidomain.AIRequestRepository
	wordRepo    WordAccessor
	contribRepo ContributionReader
	auditLog    AuditLogWriter
	llm         aidomain.LLMClient
	model       string
}

func NewEnrichmentService(
	aiRepo aidomain.AIRequestRepository,
	wordRepo WordAccessor,
	contribRepo ContributionReader,
	auditLog AuditLogWriter,
	llm aidomain.LLMClient,
	model string,
) *EnrichmentService {
	return &EnrichmentService{
		aiRepo:      aiRepo,
		wordRepo:    wordRepo,
		contribRepo: contribRepo,
		auditLog:    auditLog,
		llm:         llm,
		model:       model,
	}
}

func (s *EnrichmentService) TriggerDefinitionEnrichment(ctx context.Context, adminID, wordID uuid.UUID) (*aidomain.AIRequest, error) {
	word, err := s.wordRepo.FindByID(ctx, wordID)
	if err != nil {
		return nil, ErrWordNotFound
	}

	prompt := buildDefinitionPrompt(word)
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeEnrichDefinition, &wordID, nil, adminID, s.model, prompt)
	if err := s.aiRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	go s.runLLM(req, prompt)
	return req, nil
}

func (s *EnrichmentService) TriggerExampleSuggestion(ctx context.Context, adminID, wordID uuid.UUID) (*aidomain.AIRequest, error) {
	word, err := s.wordRepo.FindByID(ctx, wordID)
	if err != nil {
		return nil, ErrWordNotFound
	}

	prompt := buildExamplePrompt(word)
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeSuggestExample, &wordID, nil, adminID, s.model, prompt)
	if err := s.aiRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	go s.runLLM(req, prompt)
	return req, nil
}

func (s *EnrichmentService) TriggerRelatedWordSuggestion(ctx context.Context, adminID, wordID uuid.UUID) (*aidomain.AIRequest, error) {
	word, err := s.wordRepo.FindByID(ctx, wordID)
	if err != nil {
		return nil, ErrWordNotFound
	}

	prompt := buildRelatedPrompt(word)
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeSuggestRelated, &wordID, nil, adminID, s.model, prompt)
	if err := s.aiRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	go s.runLLM(req, prompt)
	return req, nil
}

func (s *EnrichmentService) TriggerQualityCheck(ctx context.Context, adminID, contributionID uuid.UUID) (*aidomain.AIRequest, error) {
	contrib, err := s.contribRepo.FindByID(ctx, contributionID)
	if err != nil {
		return nil, ErrContributionNotFound
	}

	prompt := buildQualityCheckPrompt(contrib)
	req := aidomain.NewAIRequest(aidomain.AIRequestTypeQualityCheck, nil, &contributionID, adminID, s.model, prompt)
	if err := s.aiRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	go s.runLLM(req, prompt)
	return req, nil
}

func (s *EnrichmentService) ApproveAIRequest(ctx context.Context, adminID, requestID uuid.UUID) (*aidomain.AIRequest, error) {
	req, err := s.aiRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, ErrAIRequestNotFound
	}

	if err := req.Approve(adminID); err != nil {
		if errors.Is(err, aidomain.ErrCannotApproveQualityCheck) || errors.Is(err, aidomain.ErrAIInvalidTransition) {
			return nil, ErrAIRequestConflict
		}
		return nil, err
	}

	if req.TargetWordID != nil {
		if mergeErr := s.mergeAIOutput(ctx, req); mergeErr != nil {
			return nil, mergeErr
		}
	}

	if err := s.aiRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	auditLog, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionApproveAI, "ai_request", requestID, nil)
	_ = s.auditLog.CreateAuditLog(ctx, auditLog)

	return req, nil
}

func (s *EnrichmentService) RejectAIRequest(ctx context.Context, adminID, requestID uuid.UUID) (*aidomain.AIRequest, error) {
	req, err := s.aiRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, ErrAIRequestNotFound
	}

	if err := req.Reject(adminID); err != nil {
		if errors.Is(err, aidomain.ErrAIInvalidTransition) {
			return nil, ErrAIRequestConflict
		}
		return nil, err
	}

	if err := s.aiRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	auditLog, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionRejectAI, "ai_request", requestID, nil)
	_ = s.auditLog.CreateAuditLog(ctx, auditLog)

	return req, nil
}

func (s *EnrichmentService) ListByWord(ctx context.Context, wordID uuid.UUID, page, perPage int) ([]*aidomain.AIRequest, int, error) {
	return s.aiRepo.ListByWord(ctx, wordID, page, perPage)
}

func (s *EnrichmentService) ListPendingReview(ctx context.Context, page, perPage int) ([]*aidomain.AIRequest, int, error) {
	return s.aiRepo.ListPendingReview(ctx, page, perPage)
}

// ─── LLM runner ──────────────────────────────────────────────────────────────

func (s *EnrichmentService) runLLM(req *aidomain.AIRequest, prompt string) {
	ctx := context.Background()
	resp, err := s.llm.Complete(ctx, aidomain.CompletionRequest{
		Model: s.model,
		Messages: []aidomain.Message{
			{Role: "system", Content: "You are a Banjar Hulu language expert. Respond ONLY with valid JSON, no markdown."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
	})

	rawResp := map[string]any{"model": s.model}
	if err != nil {
		rawResp["error"] = err.Error()
		_ = req.MarkFailed(rawResp)
	} else {
		rawResp["content"] = resp.Content
		rawResp["model"] = resp.Model
		parsed := extractJSON(resp.Content)
		_ = req.MarkCompleted(rawResp, parsed)
	}

	_ = s.aiRepo.Update(ctx, req)
}

// ─── output merger ────────────────────────────────────────────────────────────

func (s *EnrichmentService) mergeAIOutput(ctx context.Context, req *aidomain.AIRequest) error {
	if req.ParsedOutput == nil {
		return nil
	}
	word, err := s.wordRepo.FindByID(ctx, *req.TargetWordID)
	if err != nil {
		return ErrWordNotFound
	}

	switch req.Type {
	case aidomain.AIRequestTypeEnrichDefinition:
		def, _ := req.ParsedOutput["definition"].(string)
		if def != "" {
			word.Source = dictdomain.SourceAIGenerated
			if err := word.AddDefinition(def, len(word.Definitions)+1); err != nil {
				return err
			}
		}

	case aidomain.AIRequestTypeSuggestExample:
		banjar, _ := req.ParsedOutput["banjar_sentence"].(string)
		indonesian, _ := req.ParsedOutput["indonesian_translation"].(string)
		if banjar != "" {
			word.Source = dictdomain.SourceAIGenerated
			word.AddExample(banjar, indonesian)
		}

	case aidomain.AIRequestTypeSuggestRelated:
		relBanjar, _ := req.ParsedOutput["related_word"].(string)
		if relBanjar != "" {
			related, err := s.wordRepo.FindByBanjar(ctx, relBanjar)
			if err == nil && related != nil {
				word.AddRelatedWord(related.ID)
			}
		}

	case aidomain.AIRequestTypeQualityCheck:
		// quality_check results are informational only; no word mutation needed.
	}

	return s.wordRepo.Update(ctx, word)
}

// ─── prompt builders ──────────────────────────────────────────────────────────

func buildDefinitionPrompt(word *dictdomain.Word) string {
	defs := make([]string, len(word.Definitions))
	for i, d := range word.Definitions {
		defs[i] = fmt.Sprintf("%d. %s", i+1, d.Meaning)
	}
	existing := "none"
	if len(defs) > 0 {
		existing = strings.Join(defs, "; ")
	}
	return fmt.Sprintf(
		`Banjar Hulu word: "%s" (word class: %s).
Existing definitions: %s.
Suggest one additional Indonesian definition that is not already listed.
Respond ONLY with JSON: {"definition":"<Indonesian definition>"}`,
		word.Banjar, word.WordClass, existing,
	)
}

func buildExamplePrompt(word *dictdomain.Word) string {
	defs := make([]string, len(word.Definitions))
	for i, d := range word.Definitions {
		defs[i] = d.Meaning
	}
	meaning := "unknown"
	if len(defs) > 0 {
		meaning = defs[0]
	}
	return fmt.Sprintf(
		`Banjar Hulu word: "%s" meaning: "%s".
Write one natural example sentence using this word in Banjar Hulu dialect with its Indonesian translation.
Respond ONLY with JSON: {"banjar_sentence":"<Banjar sentence>","indonesian_translation":"<Indonesian translation>"}`,
		word.Banjar, meaning,
	)
}

func buildRelatedPrompt(word *dictdomain.Word) string {
	return fmt.Sprintf(
		`Banjar Hulu word: "%s" (word class: %s).
Suggest the single most related Banjar Hulu word (synonym, antonym, or closely related concept).
Respond ONLY with JSON: {"related_word":"<Banjar word>","relationship":"<synonym|antonym|related>"}`,
		word.Banjar, word.WordClass,
	)
}

func buildQualityCheckPrompt(contrib *communitydomain.Contribution) string {
	payload, _ := json.Marshal(contrib.Payload)
	return fmt.Sprintf(
		`Evaluate this Banjar Hulu dictionary contribution for quality.
Payload: %s
Check: accuracy of Indonesian definitions, consistency with Banjar Hulu dialect, spelling.
Respond ONLY with JSON: {"score":<0-100>,"notes":"<evaluation>","issues":[]}`,
		string(payload),
	)
}

func extractJSON(content string) map[string]any {
	if idx := strings.Index(content, "{"); idx >= 0 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "}"); idx >= 0 {
		content = content[:idx+1]
	}
	var result map[string]any
	_ = json.Unmarshal([]byte(content), &result)
	return result
}
