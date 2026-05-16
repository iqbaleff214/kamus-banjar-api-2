package commands_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
)

// ─── fake LLM client ─────────────────────────────────────────────────────────

type fakeLLM struct {
	called   bool
	response *domain.CompletionResponse
	err      error
}

func (f *fakeLLM) Complete(_ context.Context, _ domain.CompletionRequest) (*domain.CompletionResponse, error) {
	f.called = true
	return f.response, f.err
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestTranslateText_EmptyText(t *testing.T) {
	llm := &fakeLLM{}
	svc := commands.NewTranslateService(llm, "model")
	_, err := svc.TranslateText(context.Background(), "", "")
	assert.ErrorIs(t, err, commands.ErrEmptyText)
	assert.False(t, llm.called)
}

func TestTranslateText_InputTooLong(t *testing.T) {
	llm := &fakeLLM{}
	svc := commands.NewTranslateService(llm, "model")
	long := strings.Repeat("a", 1001)
	_, err := svc.TranslateText(context.Background(), long, "")
	assert.ErrorIs(t, err, commands.ErrTextTooLong)
	assert.False(t, llm.called)
}

func TestTranslateText_Success(t *testing.T) {
	llm := &fakeLLM{
		response: &domain.CompletionResponse{
			Model:   "mistralai/mistral-7b-instruct:free",
			Content: `{"translation":"dia tidak bisa pergi ke pasar","confidence":"high","notes":"kada=tidak"}`,
		},
	}
	svc := commands.NewTranslateService(llm, "mistralai/mistral-7b-instruct:free")
	result, err := svc.TranslateText(context.Background(), "inya kada kawa tulak ka pasar", "informal")
	require.NoError(t, err)
	assert.True(t, llm.called)
	assert.Equal(t, "inya kada kawa tulak ka pasar", result.Original)
	assert.Equal(t, "dia tidak bisa pergi ke pasar", result.Translation)
	assert.Equal(t, "high", result.Confidence)
	assert.Equal(t, "hulu", result.Dialect)
	assert.NotEmpty(t, result.Notes)
}

func TestTranslateText_LLMUnavailable(t *testing.T) {
	llm := &fakeLLM{err: domain.ErrAIUnavailable}
	svc := commands.NewTranslateService(llm, "model")
	_, err := svc.TranslateText(context.Background(), "abah", "")
	assert.True(t, errors.Is(err, domain.ErrAIUnavailable))
	assert.True(t, llm.called)
}

func TestTranslateText_ExactlyMaxLength(t *testing.T) {
	llm := &fakeLLM{
		response: &domain.CompletionResponse{Model: "m", Content: `{"translation":"ok","confidence":"low","notes":""}`},
	}
	svc := commands.NewTranslateService(llm, "m")
	text := strings.Repeat("a", 1000)
	_, err := svc.TranslateText(context.Background(), text, "")
	require.NoError(t, err)
	assert.True(t, llm.called)
}
