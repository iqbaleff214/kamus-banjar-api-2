package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
)

const maxInputLength = 1000

var (
	ErrTextTooLong = errors.New("text exceeds 1000 character limit")
	ErrEmptyText   = errors.New("text must not be empty")
)

type TranslationResult struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
	Dialect     string `json:"dialect"`
	Model       string `json:"model"`
	Confidence  string `json:"confidence"`
	Notes       string `json:"notes,omitempty"`
}

type TranslateService struct {
	llm   domain.LLMClient
	model string
}

func NewTranslateService(llm domain.LLMClient, model string) *TranslateService {
	return &TranslateService{llm: llm, model: model}
}

func (s *TranslateService) TranslateText(ctx context.Context, text, contextHint string) (*TranslationResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrEmptyText
	}
	if len([]rune(text)) > maxInputLength {
		return nil, ErrTextTooLong
	}

	system := `You are a Banjar Hulu language expert and translator.
Translate the given Banjar Hulu (Dialek Hulu) text into Indonesian.
Respond ONLY with a valid JSON object, no other text:
{"translation":"<Indonesian>","confidence":"<high|medium|low>","notes":"<brief lexical notes or empty string>"}`

	userMsg := fmt.Sprintf("Text: %s", text)
	if contextHint != "" {
		userMsg += fmt.Sprintf("\nContext: %s", contextHint)
	}

	resp, err := s.llm.Complete(ctx, domain.CompletionRequest{
		Model: s.model,
		Messages: []domain.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}

	return parseTranslationResponse(text, resp), nil
}

type llmTranslationJSON struct {
	Translation string `json:"translation"`
	Confidence  string `json:"confidence"`
	Notes       string `json:"notes"`
}

func parseTranslationResponse(original string, resp *domain.CompletionResponse) *TranslationResult {
	result := &TranslationResult{
		Original:   original,
		Dialect:    "hulu",
		Model:      resp.Model,
		Confidence: "low",
	}

	// Extract JSON from content (model may wrap it in markdown code fences)
	content := resp.Content
	if idx := strings.Index(content, "{"); idx >= 0 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "}"); idx >= 0 {
		content = content[:idx+1]
	}

	var parsed llmTranslationJSON
	if err := json.Unmarshal([]byte(content), &parsed); err == nil {
		result.Translation = parsed.Translation
		result.Notes = parsed.Notes
		switch parsed.Confidence {
		case "high", "medium", "low":
			result.Confidence = parsed.Confidence
		}
	} else {
		// Fallback: use raw content as translation
		result.Translation = strings.TrimSpace(resp.Content)
	}

	return result
}
