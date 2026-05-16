package domain

import (
	"context"
	"errors"
)

var ErrAIUnavailable = errors.New("AI service unavailable")

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type CompletionResponse struct {
	Model   string
	Content string
}

type LLMClient interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}
